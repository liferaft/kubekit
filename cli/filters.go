package cli

import (
	"fmt"

	"github.com/liferaft/kubekit/pkg/kluster"
	"github.com/spf13/cobra"
)

// GetFilters returns the cluster filters to from commands such as `get clusters`.
// It returns warning or errors if the filters are not use correctly
func GetFilters(cmd *cobra.Command) (filter map[string]string, warns []string, err error) {
	warns = make([]string, 0)
	filter = make(map[string]string, 0)

	filterFlag := cmd.Flags().Lookup("filter")
	if filterFlag == nil {
		return filter, warns, nil
	}

	filterStr := filterFlag.Value.String()
	filter, orphanKeys, err := StringToMap(filterStr)
	if err != nil {
		return filter, warns, err
	}

	if len(orphanKeys) != 0 {
		var plural string
		if len(orphanKeys) > 1 {
			plural = "s"
		}
		warn := fmt.Sprintf("incorrect use of filter%s %s, use them in this form: --filter parameter1=value1 --filter parameter2=value2", plural, orphanKeys)
		warns = append(warns, warn)
	}

	if invalidFilterFields := kluster.InvalidFilterParams(filter); len(invalidFilterFields) != 0 {
		var plural string
		if len(invalidFilterFields) > 1 {
			plural = "s"
		}
		warn := fmt.Sprintf("invalid filter%s: %s", plural, invalidFilterFields)
		warns = append(warns, warn)
	}

	return filter, warns, nil
}
