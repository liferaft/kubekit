# Make coding more python3-ish

from __future__ import (absolute_import, division, print_function)
__metaclass__ = type

DOCUMENTATION = '''
  callback: kubekit
  type: stdout
  short_description: Ansible screen and file output for KubeKit
  version_added: "2.5"
  description:
    - Append the tasks result to a file in yaml format
    - Send the default output to a log file
    - Prints in JSON the final stats in one line to stdout
  extends_documentation_fragment:
    - default_callback
  requirements:
    - set as stdout in configuation
'''

import yaml
import json
import os


from datetime import datetime
from threading import Thread, Timer, Lock
from flask import Flask, make_response
from werkzeug.serving import make_server

# try:
#     from flask import Flask, make_response
#     from werkzeug.serving import make_server
#     HAS_FLASK = True
# except ImportError:
#     HAS_FLASK = False

from ansible.constants import TREE_DIR
from ansible.utils.path import makedirs_safe
from ansible.module_utils._text import to_bytes
from ansible.parsing.yaml.dumper import AnsibleDumper
from ansible.plugins.callback import CallbackBase, strip_internal_keys
from ansible.plugins.callback.default import CallbackModule as CallbackModule_default

threadLockResults = Lock()
threadLockStats = Lock()

app = Flask('kubekit-configurator')

DEFAULT_PORT = 1080
ENV_VAR_PORT = 'KUBEKIT_CONFIGURATOR_PORT'
DEFAULT_EXTRA_TIME = 1 * 60 # 1 minutes
ENV_VAR_EXTRA_TIME = 'KUBEKIT_CONFIGURATOR_ETIME'

class ServerThread(Thread):
  def __init__(self, app):
    Thread.__init__(self)

    # last_results_data is flushed everytime the server provide the results
    self.last_results_data = []
    # results_data store all the results
    self.results_data = []
    self.stats_data = {}

    self.port = os.getenv(ENV_VAR_PORT, DEFAULT_PORT)
    self.srv = make_server('0.0.0.0', self.port, app)
    self.ctx = app.app_context()
    self.ctx.push()

  def run(self):
    self.srv.serve_forever()
  
  def shutdown(self):
    self.srv.shutdown()

  def append_task(self, task):
    threadLockResults.acquire()
    self.last_results_data.append(task)
    self.results_data.append(task)
    threadLockResults.release()

  def get_tasks(self):
    threadLockResults.acquire()
    results = self.last_results_data
    self.last_results_data = []
    threadLockResults.release()
    return results

  def set_stats(self, stats):
    threadLockStats.acquire()
    self.stats_data = stats
    threadLockStats.release()

  def get_stats(self):
    threadLockStats.acquire()
    stats = self.stats_data
    threadLockStats.release()
    return stats


def make_json_response(content):
  json_content = json.dumps(content, sort_keys=True, indent=4)
  response = make_response(json_content)
  response.headers['Content-type'] = "application/json"
  return response

@app.route("/tasks", methods=['GET'])
def lastest_tasks():
  global server
  return make_json_response(server.get_tasks())

@app.route("/results", methods=['GET'])
def results():
  global server
  return make_json_response(server.results_data)

@app.route("/stats", methods=['GET'])
def stats():
  global server
  return make_json_response(server.get_stats())

@app.route("/shutdown", methods=['GET'])
def stop_server():
  global server
  server.shutdown()


class CallbackModule(CallbackModule_default):  # pylint: disable=too-few-public-methods,no-init
  '''
  Override for the default callback module.

  Render std err/out outside of the rest of the result which it prints with
  indentation.
  '''

  CALLBACK_VERSION = 2.0
  CALLBACK_TYPE = 'stdout'
  CALLBACK_NAME = 'kubekit'

  STATUS_FAILED = 'failed'
  STATUS_OK     = 'ok'

  def __init__(self):
    self.super_ref = super(CallbackModule, self)
    self.super_ref.__init__()

    # if not HAS_FLASK:
    #   self.disabled = True
    #   self._display.warning("The required Flask is not installed. "
    #                         "pip install Flask")

    self.playbook_data = {}

    self.results_data = []
    self.stats_data = {}

    self.last_task = {}

    self.errors = 0
    self.start_time = datetime.utcnow()

    self.etime = os.getenv(ENV_VAR_EXTRA_TIME, DEFAULT_EXTRA_TIME)

    global server

    server = ServerThread(app)
    server.start()
    self._display.warning("server started on port %s" % server.port)


    # dir = TREE_DIR
    # if not dir:
    #   dir = os.path.join(os.path.sep, 'var', 'log', 'kubekit')
    #   self._display.warning("The kubekit callback is defaulting to %s, as an invalid directory was provided. %s" % (dir, TREE_DIR))
    # makedirs_safe(dir)
    # self.path = os.path.join(dir, 'configurator.yaml')
    # try:
    #   os.remove(self.path)
    # except (OSError, IOError) as e:
    #   self._display.warning("Unable to remove file %s: %s" % (self.path, str(e)))
    # self.task_data = {}
    # self.last_task = None
    # self.shown_title = False
    # self.writeToFile('results:\n')

  # def writeToFile(self, buf):
  #   buf = to_bytes(buf)
  #   try:
  #     with open(self.path, 'a+') as fd:
  #       fd.write(buf)
  #   except (OSError, IOError) as e:
  #     self._display.warning("Unable to write to file %s: %s" % (self.path, str(e)))

  def new_task(self, task):
    self.append_task()
    self.last_task = {
        'name': task.get_name().strip(),
        'uuid': task._uuid,
    }

  def update_task(self, result, status=STATUS_OK):
    self.last_task['status'] = status
    self.last_task['changed'] = result._result.get('changed', False)
    self.last_task['node'] = result._host.get_name()
    # self.last_task['result'] = self._dump_results(result._result)

  def append_task(self):
    global server

    if self.last_task:
      self.results_data.append(self.last_task)
      server.append_task(self.last_task)
      self.last_task = {}

  def process_results(self, result, status=STATUS_OK):
    if self.last_task['uuid'] != result._task._uuid:
      self.new_task(result._task)
    if status != self.STATUS_OK:
      self.errors += 1
    self.update_task(result, status)
    self.append_task()

  def process_item_results(self, result, status=STATUS_OK):
    # If there isn't a status, assign it
    if self.last_task.get('status') == None:
      self.last_task['status'] = status
    # If the status is OK but this is a Failure (or not OK), assign it. 
    # Or, if it's not OK, keep it
    if self.last_task['status'] == self.STATUS_OK and status != self.STATUS_OK:
      self.last_task['status'] = status

    if status != self.STATUS_OK:
      self.errors += 1

    item_data = {
        'name': "%s" % (self._get_item(result._result),),
        'status': status,
        'changed': result._result.get('changed', False),
        'node': result._host.get_name(),
    }
    
    if self.last_task.get('items') == None:
      self.last_task['items'] = [ item_data ]
    else:
      self.last_task['items'].append(item_data)

  def process_stats(self, stats):
    global server

    end_time = datetime.utcnow()
    runtime = end_time - self.start_time
    status = self.STATUS_OK if self.errors == 0 else self.STATUS_FAILED

    hosts = sorted(stats.processed.keys())

    summarized_stats = {}
    for h in hosts:
      summarized_stats[h] = stats.summarize(h)
    
    self.stats_data = {
      'status': status,
      'duration': runtime.total_seconds(),
      'stats': summarized_stats,
    }
    server.set_stats(self.stats_data)

  def process_playbook_data(self):
    global server

    self.playbook_data = {
      'results': self.results_data,
      'stats': self.stats_data,
    }
    self._display.warning("shutting down kubekit configurator server in %s seconds" % self.etime)
    Timer(self.etime, server.shutdown).start()


  def v2_runner_on_failed(self, result, ignore_errors=False):
    self.process_results(result, self.STATUS_FAILED)
    self.super_ref.v2_runner_on_failed(result, ignore_errors)

  def v2_runner_on_ok(self, result):
    self.process_results(result)
    self.super_ref.v2_runner_on_ok(result)

  def v2_runner_on_skipped(self, result):
    self.process_results(result)
    self.super_ref.v2_runner_on_skipped(result)

  # this won't happen because Ansible only runs locally
  # def v2_runner_on_unreachable(self, result):

  # this won't happen because Ansible only runs locally
  # def v2_playbook_on_no_hosts_matched(self):

  # this won't happen because Ansible only runs locally
  # def v2_playbook_on_no_hosts_remaining(self):

  def v2_playbook_on_task_start(self, task, is_conditional):
    self.new_task(task)
    self.super_ref.v2_playbook_on_task_start(task, is_conditional)

  # this won't happen because don't have these tasks
  # def v2_playbook_on_cleanup_task_start(self, task):

  # do default or pass
  # def v2_playbook_on_handler_task_start(self, task):

  # do default or pass
  # def v2_playbook_on_play_start(self, play):
  #   self.super_ref.v2_playbook_on_play_start(play)

  # do default
  # def v2_on_file_diff(self, result):

  # Old definition in v2.0
  def v2_playbook_item_on_ok(self, result):
      self.v2_runner_item_on_ok(result)

  def v2_runner_item_on_ok(self, result):
    self.process_item_results(result)
    self.super_ref.v2_runner_item_on_ok(result)

  # Old definition in v2.0
  def v2_playbook_item_on_failed(self, result):
      self.v2_runner_item_on_failed(result)

  def v2_runner_item_on_failed(self, result):
    self.process_item_results(result, self.STATUS_FAILED)
    self.super_ref.v2_runner_item_on_failed(result)

  # Old definition in v2.0
  def v2_playbook_item_on_skipped(self, result):
      self.v2_runner_item_on_skipped(result)

  def v2_runner_item_on_skipped(self, result):
    self.process_item_results(result)
    self.super_ref.v2_runner_item_on_skipped(result)

  # do default or pass
  # def v2_playbook_on_include(self, included_file):

  def v2_playbook_on_stats(self, stats):
    self.process_stats(stats)
    self.process_playbook_data()
    self.super_ref.v2_playbook_on_stats(stats)

  # do default or pass
  # def v2_playbook_on_start(self, playbook):
  #   self.super_ref.v2_playbook_on_start(playbook)

  # do default
  # def v2_runner_retry(self, result):

  # do default
  # def v2_playbook_on_notify(self, handler, host):
