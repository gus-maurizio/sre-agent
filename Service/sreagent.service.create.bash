#!/usr/bin/env bash

touch /etc/systemd/system/sreagent.service
chmod 664 /etc/systemd/system/sreagent.service
mkdir -p /opt/sreagent

if [ $(getent group sreagent) ]; then
  echo "group exists."
else
  echo "group does not exist."
  groupadd sreagent
fi

if [ $(getent passwd sreagent) ]; then
  echo "user exists."
else
  echo "user does not exist."
  useradd --no-create-home --gid sreagent sreagent
fi

tar -xvf sreagent.Linux.tgz --directory /opt/sreagent
mkdir -p /opt/sreagent/log /opt/sreagent/run
chown -R sreagent:sreagent /opt/sreagent
usermod --home=/opt/sreagent sreagent
cp etc_systemd_system_sreagent.service /etc/systemd/system/sreagent.service
chmod 664 /etc/systemd/system/sreagent.service
# Try to find if a real system or a container (no systemctl)
if [ $(which systemctl) ]; then
  echo We found systemctl
  sudo systemctl start  sreagent
  sudo systemctl status sreagent
  sudo systemctl enable sreagent
else
  echo We did not find systemctl, container?
  su - sreagent -c "nohup /opt/sreagent/bin/sre-agent -f /opt/sreagent/config/agent.yaml >> /opt/sreagent/log/daemon.stdout.log 2>/opt/sreagent/log/daemon.stderr.log &"
  ls -lhrt /opt/sreagent/run
fi

