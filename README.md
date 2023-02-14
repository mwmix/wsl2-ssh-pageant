# WSL2GPGGO
Inspired by the now or soon to be archived project [wsl2-ssh-pagent](https://github.com/BlackReloaded/wsl2-ssh-pageant).

# TODO
Original version used socat exec maybe make this not require that.
Just allow this thing to run on it's own indefinitely.

# Maybe?
Place in /home/matt/.gnupg/S.gpg-agent
```
%Assuan%
socket=/home/matt/temp/WSL2GPGGo/gpg.sock
```
If I can get that socket thing working I can probably also have it point to my windows path C:\Users\Matt\AppData\Local\gnupg\S.gpg-agent ??

# Stuff
```bash
# Removing Linux SSH socket and replacing it by link to wsl2-ssh pageant socket
export SSH_AUTH_SOCK="$HOME/.ssh/agent.sock"
if ! ss -a | grep -q "$SSH_AUTH_SOCK"; then
  rm -f "$SSH_AUTH_SOCK"
  wsl2_ssh_pageant_bin="/mnt/c/Users/matt/.ssh/wsl2-ssh-pageant.exe"
  if test -x "$wsl2_ssh_pageant_bin"; then
    (setsid nohup socat UNIX-LISTEN:"$SSH_AUTH_SOCK,fork" EXEC:"$wsl2_ssh_pageant_bin" >/dev/null 2>&1 &)
  else
    echo >&2 "WARNING: $wsl2_ssh_pageant_bin is not executable."
  fi
  unset wsl2_ssh_pageant_bin
fi
# GPG Socket
# Removing Linux GPG Agent socket and replacing it by link to wsl2-ssh-pageant GPG socket
export GPG_AGENT_SOCK="$HOME/.gnupg/S.gpg-agent"
if ! ss -a | grep -q "$GPG_AGENT_SOCK"; then
  rm -rf "$GPG_AGENT_SOCK"
  wsl2_ssh_pageant_bin="/mnt/c/Users/matt/.ssh/wsl2-ssh-pageant.exe"
  if test -x "$wsl2_ssh_pageant_bin"; then
    (setsid nohup socat UNIX-LISTEN:"$GPG_AGENT_SOCK,fork" EXEC:"$wsl2_ssh_pageant_bin -verbose -logfile /home/matt/wsl2log.txt -gpgConfigBasepath /mnt/c/Users/matt/AppData/Local/gnupg --gpg S.gpg-agent" >/dev/null 2>&1 &)
  else
    echo >&2 "WARNING: $wsl2_ssh_pageant_bin is not executable."
  fi
  unset wsl2_ssh_pageant_bin
fi
```

# Useful Reference Material
- https://www.gnupg.org/documentation/manuals/assuan/Socket-wrappers.html
- https://www.gnupg.org/documentation/manuals/gnupg/Agent-Protocol.html
- https://www.gnupg.org/documentation/manuals/gnupg/Invoking-GPG_002dAGENT.html#Invoking-GPG_002dAGENT
- https://www.gnupg.org/documentation/manuals/gnupg/Option-Index.html#Option-Index
- https://github.com/gpg/gnupg/blob/master/tools/gpg-connect-agent.c