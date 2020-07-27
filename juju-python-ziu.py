import os
import sys
import time
import subprocess
from datetime import datetime
minutes_verify = 30

def verify_deployment(minutes):
    tries = 0
    status_verified = False
    while (status_verified is False and tries<minutes):
        tries += 1
        time.sleep(60)
        c1 = subprocess.Popen(["juju", "status"], stdout=subprocess.PIPE)
        c2 = subprocess.Popen(["grep", "-e allocating", "-e blocked", "-e pending", "-e waiting", "-e maintenance", "-e executing", "-e error"],
                stdin=c1.stdout,stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        stdout = c2.communicate()[0]
        if not stdout:
            status_verified = True
        c1.wait()
    if status_verified is False:
        return -1,"Tries Expired, ",tries
    else:
        return 0,"",tries


def verify_upgrade():
    tries = 0
    status_verified = False
    while (status_verified is False and tries<minutes_verify):
        tries += 1
        stages = 0
        completed = 0
        time.sleep(60)
        c1 = subprocess.Popen(["juju", "status"], stdout=subprocess.PIPE)
        c2 = subprocess.Popen(["grep", "stage/done"], stdin=c1.stdout, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        stdout = c2.communicate()[0]
        if stdout:
            lines = stdout.split("\n")
            for item in lines:
                if item != "":
                    stages += 1
                    if item.find("5/5") != -1:
                        completed += 1
            if stages == completed:
                status_verified = True
        c1.wait()
    if status_verified is False:
        return -1,"Tries Expired, ",tries
    else:
        return 0,"",tries

def upgrade_procedure():
    controller = subprocess.call(["./controller-upgrade.sh", sys.argv[1]], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    return_code,return_message,time = verify_upgrade()
    if return_code == -1:
        return_message += "Controller Upgrade Failed in approx ~" + str(time) + " minutes\n"
    else:
        return_message += "Controller Upgrade Successful in approx ~" + str(time) + " minutes\n"
        agents = subprocess.call(["./agent-upgrade.sh"], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        rc,rm,t =  verify_deployment(30)
        if rc == -1:
            return_message += rm + "Computes Upgrade Failed in approx ~" + str(t) + " minutes\n"
        else:
            return_message += rm + "Computes Upgrade Successful in approx ~" + str(t) + " minutes\n"
    return return_message


def zero_impact_upgrade():
    global start,end,message
    start = datetime.now()
    return_code,return_message,_ = verify_deployment(1)
    if return_code == -1:
        message = "\nPreviously faulty deployment, Upgrade not possible"
    else:
        upgrade_result = upgrade_procedure()
        message = return_message + upgrade_result
    end = datetime.now()


def write_result():
    result = "\nJuju Zero Impact Upgrade to " + str(sys.argv[1]) + "\nStarted at " + str(start) + "\nEnded at " + str(end) + "\nTime taken = " + str(end-start) + "\n" + message
    fout = "result.txt"
    if os.path.exists(fout):
        os.remove(fout)
    with open(fout, 'w') as outfile:
        outfile.write(result)

def main():
    zero_impact_upgrade()
    write_result()

if __name__ == "__main__":
    main()
