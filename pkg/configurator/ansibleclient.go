package configurator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/johandry/log"
	"github.com/kraken/ui"
)

// Ansible status possible values
const (
	AnsibleStatusOk     = "ok"
	AnsibleStatusFailed = "failed"
)

// KubeKitConfiguratorPort is the port used by the KubeKit Ansible callback to
// expose the Ansible playbook logs
const KubeKitConfiguratorPort = 1080

// Maximum number retries and frecuency to get tasks and status
const (
	MaxEmptyTaskRetries   = 5
	MaxFailedTaskRetries  = 10
	MaxFailedStatsRetries = 5

	SecondsTickTasks = 5
	SecondsTickStats = 10
)

// AnsibleTaskItem contain the results of the items in a tasks (if any)
type AnsibleTaskItem struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Node    string `json:"node"`
}

func (i AnsibleTaskItem) String() string {
	var changedMsg string
	if i.Changed {
		changedMsg = "(changed)"
	}
	// Remove the anoying: u'string'
	name := strings.Replace(i.Name, "u'", "'", -1)
	return fmt.Sprintf("[%s]      ITEM %q %s", i.Node, name, changedMsg)
}

// Ok returns true if the status is OK
func (i AnsibleTaskItem) Ok() bool {
	if len(i.Status) == 0 {
		return true
	}
	return i.Status == AnsibleStatusOk
}

// Log print the task item to a logger
func (i AnsibleTaskItem) Log(logger *log.Logger) {
	if i.Ok() {
		logger.Infof("%s", i)
	} else {
		logger.Errorf("%s", i)
	}
}

// AnsibleTask contain the results of a task execution
type AnsibleTask struct {
	Name    string            `json:"name"`
	UUID    string            `json:"uuid"`
	Status  string            `json:"status"`
	Changed bool              `json:"changed"`
	Node    string            `json:"node"`
	Items   []AnsibleTaskItem `json:"items"`
}

func (t AnsibleTask) String() string {
	var changedMsg string
	if t.Changed {
		changedMsg = "(changed)"
	}

	return fmt.Sprintf("[%s] TASK %q %s", t.Node, t.Name, changedMsg)
}

// Ok returns true if the status is OK
func (t AnsibleTask) Ok() bool {
	if len(t.Status) == 0 {
		return true
	}
	return t.Status == AnsibleStatusOk
}

// Report prints the task to the UI and Logs
func (t AnsibleTask) Report(uI *ui.UI) {
	if t.Ok() {
		uI.Log.Infof(t.String())
	} else {
		uI.Log.Errorf(t.String())
	}

	taskNames := strings.Split(t.Name, " : ")
	if len(taskNames) != 2 {
		return
	}

	if !t.Ok() {
		// If a task fail (not an item, tasks contain " : "), report it as complete
		uI.Notify(t.Node, ui.Red+taskNames[0], "</"+taskNames[0]+">", "")
	}

	// Only prints the tasks with tags
	if strings.Trim(taskNames[1], "</>") != taskNames[0] {
		return
	}

	color := ui.Green
	if !t.Ok() {
		color = ui.Red
	}
	task := fmt.Sprintf("%s%s", color, taskNames[0])

	uI.Notify(t.Node, task, taskNames[1], "", ui.Configure)
}

// AnsibleHostStat contain the Ansible stats per host
type AnsibleHostStat struct {
	Changed     int `json:"changed"`
	Failures    int `json:"failures"`
	Ok          int `json:"ok"`
	Skipped     int `json:"skipped"`
	Unreachable int `json:"unreachable"`
}

// Log print the stats of a host to a logger. Really rare situation, but here
// just in case
func (hstats AnsibleHostStat) Log(host string, duration float32, logger *log.Logger) {
	message := fmt.Sprintf("[%s] STATS: ok=%d    changed=%d    failed=%d    unreachable=%d    skipped=%d    total duration=%gs", host, hstats.Ok, hstats.Changed, hstats.Failures, hstats.Unreachable, hstats.Skipped, duration)
	if hstats.Failures > 0 || hstats.Unreachable > 0 {
		logger.Error(message)
	} else {
		logger.Info(message)
	}
}

// AnsibleStats encapsulate all the stats
type AnsibleStats struct {
	Duration float32                    `json:"duration"`
	Status   string                     `json:"status"`
	Stats    map[string]AnsibleHostStat `json:"stats"`
}

func (stats *AnsibleStats) String() string {
	if len(stats.Stats) == 0 {
		return fmt.Sprintf("duration: %g", stats.Duration)
	}
	if len(stats.Stats) == 1 {
		for host, hostStats := range stats.Stats {
			return fmt.Sprintf("[%s] STATS: ok=%d    changed=%d    failed=%d    unreachable=%d    skipped=%d    duration=%gs", host, hostStats.Ok, hostStats.Changed, hostStats.Failures, hostStats.Unreachable, hostStats.Skipped, stats.Duration)
		}
	}
	return ""
}

// Empty returns true if this stats hasn't been collected
func (stats *AnsibleStats) Empty() bool {
	return stats.Duration == 0.0 || len(stats.Stats) == 0
}

// Ok returns true if the status is OK
func (stats *AnsibleStats) Ok() bool {
	if len(stats.Stats) == 0 {
		return false
	}
	return stats.Status == AnsibleStatusOk
}

// Log print the stats to a logger
func (stats *AnsibleStats) Log(failedTasks []AnsibleTask, logger *log.Logger) {
	if len(stats.Stats) <= 1 {
		if stats.Ok() {
			logger.Infof("%s", stats)
			return
		}
		var fTasksMsg string
		if len(failedTasks) > 0 {
			fTasksMsg = fmt.Sprintf("%s\nFailed tasks:", ui.BlackBold)
			for _, task := range failedTasks {
				fTasksMsg = fmt.Sprintf("%s\n\t[%s] FAILED TASK %q", fTasksMsg, task.Node, task.Name)
			}
		}
		logger.Errorf("%s%s", stats, fTasksMsg)
		return
	}

	for host, hostStats := range stats.Stats {
		hostStats.Log(host, stats.Duration, logger)
	}
}

// AnsibleClient handle the Ansible connection with the Ansible Callback API
type AnsibleClient struct {
	Hostname           string
	Tasks              []AnsibleTask
	Stats              *AnsibleStats
	mu                 sync.Mutex
	reqTasks           *http.Request
	reqStats           *http.Request
	retriesEmptyTasks  int
	retriesFailedTasks int
	retriesFailedStats int
	stop               *chan bool
	ui                 *ui.UI
}

func newAnsibleClient(host Host, ui *ui.UI, stop *chan bool) (*AnsibleClient, error) {
	reqTasks := getHTTPRequest(host.PublicIP, "/tasks")
	if reqTasks == nil {
		return nil, fmt.Errorf("failed to create a get request to get the ansible tasks from %s (%s)", host.RoleName, host.PublicIP)
	}

	reqStats := getHTTPRequest(host.PublicIP, "/stats")
	if reqStats == nil {
		return nil, fmt.Errorf("failed to create a get request to get the ansible stats from %s (%s)", host.RoleName, host.PublicIP)
	}

	return &AnsibleClient{
		Hostname: host.RoleName,
		reqTasks: reqTasks,
		reqStats: reqStats,
		stop:     stop,
		ui:       ui,
	}, nil
}

func (a *AnsibleClient) resetCounters() {
	a.retriesEmptyTasks = 0
	a.retriesFailedTasks = 0
	a.retriesFailedStats = 0
}

// handleError handles connection or unmarshalling errors
func (a *AnsibleClient) handleError(err error) {
	a.retriesFailedTasks++
	if a.retriesFailedTasks < MaxFailedTaskRetries {
		a.ui.Log.Warnf("[%s] failed to get the latest tasks, try %d / %d, trying again. %s", a.Hostname, a.retriesFailedTasks, MaxFailedTaskRetries, err)
		return
	}
	// Rule #2
	a.ui.Log.Errorf("[%s] failed to get tasks in the last %d times, stopping the Ansible client. %s", a.Hostname, a.retriesFailedTasks, err)
	*a.stop <- true
}

// noTasksRetreived do actions when no tasks were retreived
func (a *AnsibleClient) noTasksRetreived() *time.Ticker {
	a.retriesEmptyTasks++
	if a.retriesEmptyTasks < MaxEmptyTaskRetries {
		a.ui.Log.Debugf("[%s] got empty tasks, try %d / %d, trying again in %ds", a.Hostname, a.retriesEmptyTasks, MaxEmptyTaskRetries, SecondsTickTasks)
		return nil
	}
	// Rule #1
	a.ui.Log.Debugf("[%s] got empty tasks in the last %d times, starting the stats request", a.Hostname, a.retriesEmptyTasks)
	return time.NewTicker(SecondsTickStats * time.Second)
}

// PrintLogs prints to a given logger the ansible
func (a *AnsibleClient) PrintLogs() {
	var tickStats *time.Ticker
	var tickStatsCh <-chan time.Time
	tickTasks := time.NewTicker(SecondsTickTasks * time.Second).C

	// Workflow:
	// 1 - If there isn't any task in the latest 3 times (MaxEmptyTaskRetries), start requesting stats if you aren't
	// 2 - If failed to get task in the latest 10 times (MaxFailedTaskRetries), stop requesting everything
	// 3 - If failed to get stats in the latest 10 times, stop requesting everithing
	// 4 - If there is a retreived task and it's requesting stats, stop requesting stats
	// 5 - If there isn't any stats in the latest X times, stop requesting stats and stop

	for {
		select {
		case <-tickTasks:
			var lastestTasks []AnsibleTask
			var err error
			if lastestTasks, err = a.getLatestTasks(); err != nil {
				// Rule #2
				a.handleError(err)
				continue
			}
			if len(lastestTasks) == 0 {
				// Rule #1
				if tickStats == nil {
					if tickStats = a.noTasksRetreived(); tickStats != nil {
						tickStatsCh = tickStats.C
					}
				}
				continue
			}
			a.printTasks(lastestTasks)
			a.retriesEmptyTasks = 0
			a.retriesFailedTasks = 0
			if tickStats != nil {
				tickStats.Stop()
				tickStats = nil
			}
		case <-tickStatsCh:
			if err := a.getStats(); err == nil {
				if a.Stats != nil {
					// Get latest tasks, just in case, to append them to the
					lastestTasks, _ := a.getLatestTasks()
					if len(lastestTasks) != 0 {
						a.printTasks(lastestTasks)
					}

					failedTasks := a.getFailedTasks()
					a.Stats.Log(failedTasks, a.ui.Log)
					*a.stop <- true
				} else {
					a.ui.Log.Warnf("[%s] stats are not ready yet, will try again in %ds", a.Hostname, SecondsTickStats)
				}
			} else {
				a.retriesFailedStats++
				if a.retriesFailedStats > MaxFailedStatsRetries {
					a.ui.Log.Errorf("[%s] exceed number of tries to get stats (%d > %d), stopping the Ansible client. %s", a.Hostname, a.retriesFailedStats, MaxFailedStatsRetries, err)
					*a.stop <- true
				} else {
					a.ui.Log.Warnf("[%s] try %d / %d, trying again in %ds. %s", a.Hostname, a.retriesFailedStats, MaxFailedStatsRetries, SecondsTickStats, err)
				}
			}
		case <-*a.stop:
			return
		}
	}
}

func getHTTPRequest(ipAddress, urlPath string) *http.Request {
	baseURL := fmt.Sprintf("http://%s:%d", ipAddress, KubeKitConfiguratorPort)

	request, err := http.NewRequest(http.MethodGet, baseURL+urlPath, nil)
	if err != nil {
		return nil
	}
	request.Header.Set("User-Agent", "kubekit")

	return request
}

func unmarshall(req *http.Request, v interface{}) error {
	client := http.Client{
		Timeout: time.Second * 2,
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do the request to %s", req.URL)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read the content from request to %s", req.URL)
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json response from %s. %s", req.URL, err)
	}

	return nil
}

func (a *AnsibleClient) printTasks(tasks []AnsibleTask) {
	for _, task := range tasks {
		if len(task.Node) != 0 {
			task.Node = a.Hostname
		}
		task.Report(a.ui)

		if len(task.Items) > 0 {
			for _, item := range task.Items {
				item.Log(a.ui.Log)
			}
		}
	}
}

func (a *AnsibleClient) getStats() error {
	stats := &AnsibleStats{}
	if err := unmarshall(a.reqStats, stats); err != nil {
		return err
	}
	if !stats.Empty() {
		a.Stats = stats
	}
	return nil
}

func (a *AnsibleClient) getLatestTasks() ([]AnsibleTask, error) {
	lastestTasks := []AnsibleTask{}

	if err := unmarshall(a.reqTasks, &lastestTasks); err != nil {
		return lastestTasks, err
	}

	if len(lastestTasks) != 0 {
		a.mu.Lock()
		a.Tasks = append(a.Tasks, lastestTasks...)
		a.mu.Unlock()
	}

	return lastestTasks, nil
}

func (a *AnsibleClient) getFailedTasks() []AnsibleTask {
	fTasks := []AnsibleTask{}

	for _, task := range a.Tasks {
		if !task.Ok() {
			fTasks = append(fTasks, task)
		}
	}

	return fTasks
}
