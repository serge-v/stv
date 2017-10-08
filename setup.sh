set -x
set -e

read -p 'Upgrade [yNc] ?' answer ; [[ $answer == 'c' ]] && { echo exit; exit 1; }
[[ $answer == 'y' ]] && pacman -Syu

read -p 'Clear cache [yNc] ?' answer ; [[ $answer == 'c' ]] && { echo exit; exit 1; }
[[ $answer == 'y' ]] && pacman -Ss

systemctl disable systemd-random-seed
history -c -w
