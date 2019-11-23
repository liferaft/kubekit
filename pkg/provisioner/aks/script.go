package aks

// these customScript variables are used to switch between scripts if needed as it comes in handy for testing
var (
	customScript     = EmptyScript
	customScriptName = "EmptyScriptPlaceholder"
)

// EmptyScript is just used as a placeholder for now
// NOTE: scripts would need to be ran with /bin/sh and not /bin/bash since Azure CustomScript Extension defaults to it
// NOTE: scripts should be POSIX compliant (or at the very least work in the Dash shell of Ubuntu
const EmptyScript = `
#!/bin/sh

# metadata
DEFAULT_DEVICES='sda,sdb,sr0,'  # string should end with a comma as its used in our search
DEFAULT_EPHEMERAL_PARTITION='/dev/sdb1'
DEFAULT_EPHEMERAL_MNT_PT='/mnt'

# --- GENERATED METADATA: START ---
%s
# --- GENERATED METADATA: END ---
`
