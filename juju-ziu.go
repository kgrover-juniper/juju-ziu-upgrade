package main

import(
        "fmt"
        "os/exec"
        "os"
        "log"
        "time"
	"strconv"
	"strings"
)

var (
	start time.Time
	end time.Time
	minutes int = 30
	message string
)


func check_error(e error) {
    if e != nil {
        log.Fatal("Error- ", e)
    }
}

func verify_deployment(min int) (int, string, int) {
        tries := 0
        status_verified := false
        for (status_verified==false && tries<min){
                tries++
                time.Sleep(60*time.Second) //sleep for a minute before checking the juju status
                c2 := exec.Command("grep", "-e allocating", "-e blocked", "-e pending", "-e waiting", "-e maintenance", "-e executing", "-e error")
                c1 := exec.Command("juju", "status")
                pipe, _ := c1.StdoutPipe()
                defer pipe.Close()
                c2.Stdin = pipe
                c1.Start()
                stdout, _ := c2.Output()
                if string(stdout) == "" {
                        status_verified = true
                }
                _ = c1.Wait()
        }
        if !status_verified{
                return -1,"Tries Expired, ",tries
        } else {
                return 0,"",tries
        }
}


func verify_upgrade() (int,string,int){
        tries := 0
        status_verified := false
        for (status_verified==false && tries<=minutes){
                tries++
		stages := 0
		completed := 0
                time.Sleep(60*time.Second) //sleep for a minute before checking the juju status
		c2 := exec.Command("grep", "stage/done")
                c1 := exec.Command("juju", "status")
                pipe, _ := c1.StdoutPipe()
                defer pipe.Close()
                c2.Stdin = pipe
                c1.Start()
                stdout, _ := c2.Output()
		output := string(stdout)
                if output != "" {
			lines := strings.Split(output,"\n")
			for each:=0 ; each<len(lines); each++ {
				if(lines[each] != "") {
					stages++
					if strings.Contains(lines[each], "5/5") {
						completed++
					}
				}
			}
			if stages == completed {
				status_verified = true
			}
                }
                _ = c1.Wait()
        }
        if !status_verified{
                return -1,"Tries Expired, ",tries
        } else {
                return 0,"",tries
        }
}

func upgrade_procedure() string {
	cmd := exec.Command("/bin/sh","-x", "controller-upgrade.sh", os.Args[1])
        err := cmd.Run()
        check_error(err)
	return_code,return_message,time := verify_upgrade()
	if return_code == -1 {
		return_message += "Controller Upgrade Failed in approx ~" + strconv.Itoa(time) + " minutes\n"
	} else {
		return_message += "Controller Upgrade Successful in approx ~" + strconv.Itoa(time) + " minutes\n"
		c := exec.Command("/bin/sh","-x", "agent-upgrade.sh")
		e := c.Run()
		check_error(e)
		rc,rm,t :=  verify_deployment(30)
		if rc == -1 {
			return_message += rm + "Computes Upgrade Failed in approx ~" + strconv.Itoa(t) + " minutes\n"
		} else {
			return_message += rm + "Computes Upgrade Successful in approx ~" + strconv.Itoa(t) + " minutes\n"
		}
	}
	return return_message
}

func zero_impact_upgrade(){
	start = time.Now()
        return_code,return_message,_ := verify_deployment(1)

        if return_code == -1{
                message = "\nPreviously faulty deployment, Upgrade not possible"
        } else {
		upgrade_result := upgrade_procedure()
                message = return_message + upgrade_result
        }
	end = time.Now()
}

func write_result() {
	result := "\nJuju Zero Impact Upgrade to " + os.Args[1] + "\nStarted at " + start.String() + "\nEnded at " + end.String()
	result += "\nTime taken = " + end.Sub(start).String() + "\n" + message
	file := "result.txt"
	if _, err := os.Stat(file); !os.IsNotExist(err) {
                err_file := os.Remove(file)
                check_error(err_file)
        }
        output, err_open := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
        check_error(err_open)
        defer output.Close()
	_, err_write := fmt.Fprintln(output, result)
        check_error(err_write)
}

func main() {
	zero_impact_upgrade()
	write_result()
}
