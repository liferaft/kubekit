import re
from ansible import errors

'''
This regex pulled from Ansible code
'''

def regex(value='', pattern='', ignorecase=False, multiline=False, match_type='search'):
    ''' Expose `re` as a boolean filter using the `search` method by default.
        This is likely only useful for `search` and `match` which already
        have their own filters.
    '''
    flags = 0
    if ignorecase:
        flags |= re.I
    if multiline:
        flags |= re.M
    _re = re.compile(pattern, flags=flags)
    _bool = __builtins__.get('bool')
    return _bool(getattr(_re, match_type, 'search')(value))

def hostname_validate( value, pattern='^(([a-z][a-z0-9\-]*[a-z0-9])\.*)+([a-z0-9]|[a-z0-9][a-z0-9\-]*[a-z0-9])$'):
    return regex(value, pattern, False, False, 'search')


class TestModule(object):
    '''
    custom jinja2 test for validating hostnames
    '''
    def tests(self):
        return {
            'hostname_validate': hostname_validate
        }
