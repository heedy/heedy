# this file is processed on each interactive invocation of bash

# avoid problems with scp -- don't process the rest of the file if non-interactive
[[ $- != *i* ]] && return

#PS1="`shorthostname` \! $ "
HISTSIZE=50

alias mail=mailx
alias ls='ls --color=auto'

./.tmx

PS1='\[\e[0;32m\]\u\[\e[m\] \[\e[1;34m\]\W\[\e[m\] \[\e[1;32m\]\$\[\e[m\] \[\e[1;37m\]'
trap 'echo -ne "\e[0m"' DEBUG

export PATH=$PATH:/usr/local/go/bin
export GOPATH=$(pwd)/.go:$(pwd)/connectordb
