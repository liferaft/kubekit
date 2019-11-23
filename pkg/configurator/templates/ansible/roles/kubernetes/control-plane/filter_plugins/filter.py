#!/usr/bin/python
import re

class FilterModule(object):

    is_taint = re.compile(r'^[^:]+:(Prefer)?No(Schedule|Execute){1}[^-]$')
    is_taint_untaint = re.compile(r'^[^:]+:(Prefer)?No(Schedule|Execute){1}-?$')
    is_label = re.compile(r'^[^=]+=[^=]+$')

    def filters(self):
        return {
            'valid_taints': self.validate_taints,
            'valid_labels': self.validate_labels
        }

    def validate_taints(self, taints, untaints=False ):
        ret = []
        if untaints:
            regex = self.is_taint_untaint
        else:
            regex = self.is_taint
        for taint in taints:
            if regex.match(taint+'\n'):
                ret.append(taint)
        return ret

    def validate_labels(self, labels):
       ret = []
       for label in labels:
           if self.is_label.match(label+'\n'):
               ret.append(label)
       return ret
