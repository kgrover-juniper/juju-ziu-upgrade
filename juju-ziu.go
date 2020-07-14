package main

import(
        "fmt"
        "os/exec"
        "os"
        "bufio"
        "log"
        "time"
	"regexp"
	"strings"
)

var (
	minutes int = 60
	stage int
	upgraded int
	total_count int
	deployed int
	result string = ""
)

func check_status() {
	cmd := exec.Command("juju", "status")
        stdout, err := cmd.Output()
        check_error(err)

        file := "status.txt"
        remove_previous(file)
        recent_status, err_open := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
        check_error(err_open)
        defer recent_status.Close()
	_, err_output := fmt.Fprintln(recent_status, string(stdout))
        check_error(err_output)

}

func verify_deployment(mins int) (int,string){
        tries := 0
        status_verified := false
        for (status_verified==false && tries<=mins){
                tries++
                total_count = 0
                deployed = 0
                time.Sleep(60*time.Second) //sleep for a minute before checking the juju status
                check_status()//check status after a minute
                fin, err_open_file := os.Open("status.txt")
                check_error(err_open_file)
                defer fin.Close()
                scanner := bufio.NewScanner(fin)
                for scanner.Scan() {
                        total_count += 1
                        line := scanner.Text()
                        in_progress := regexp.MustCompile(`allocating|pending|waiting|blocked|executing|maintenance`)
                        in_error := regexp.MustCompile(`error`)
                        if in_error.MatchString(line)==true {
                                return -1, "Error, "
                        }
                        if in_progress.MatchString(line)==false {
                                deployed += 1
                        }
                }
                if total_count == deployed {
                        status_verified = true
                }
        }
        if !status_verified{
                return -1,"Tries Expired, "
        } else {
                return 0,""
        }
}

func check_error(e error) {
    if e != nil {
        log.Fatal("Error- ", e)
    }
}

func remove_previous(filename string) {
	if file_exists(filename) {
		err_file := os.Remove(filename)
                check_error(err_file)
        }
}

func file_exists(fileName string) (bool) {
    _, err := os.Stat(fileName)
    if os.IsNotExist(err) {
        return false
    } else {
            return true
    }
}

func verify_upgrade() (int,string){
        tries := 0
        status_verified := false
        for (status_verified==false && tries<=minutes){
                tries++
                stage = 0
                upgraded = 0
                time.Sleep(60*time.Second) //sleep for a minute before checking the juju status
                check_status()//check status after a minute
                fin, err_open_file := os.Open("status.txt")
                check_error(err_open_file)
                defer fin.Close()
                scanner := bufio.NewScanner(fin)
                for scanner.Scan() {
                        line := scanner.Text()
			if strings.Contains(line,"stage/done = "){
				stage += 1
				words := strings.FieldsFunc(line, func(c rune) bool {
                                        return c == ' '
                                })
				for i:=0;i<len(words);i++{
                                        if (words[i]=="=" && words[i+1]=="5/5"){
                                                upgraded += 1
                                                break
                                        }
                                }
			}
                }
                if stage == upgraded {
                        status_verified = true
                }
        }
        if !status_verified{
                return -1,"Tries Expired, "
        } else {
                return 0,""
        }
}

func upgrade_procedure() string {
	cmd := exec.Command("/bin/sh","-x", "controller-upgrade.sh", os.Args[1])
        err := cmd.Run()
        check_error(err)
	return_code,return_message := verify_upgrade()
	if return_code == -1 {
		return_message += " Controller Upgrade Failed\n"
	} else {
		c := exec.Command("/bin/sh","-x", "agent-upgrade.sh")
		e := c.Run()
		check_error(e)
		rc,rm :=  verify_deployment(60)
		if rc == -1 {
			return_message += rm + " Computes Upgrade Failed\n"
		} else {
			return_message += rm + " All upgrade successful\n"
		}
	}
	return return_message
}

func zero_impact_upgrade(){
	start := time.Now()
        return_code,return_message := verify_deployment(1)

        if return_code == -1{
                return_message += "\n Previously faulty deployment, Upgrade not possible"
        } else {
		upgrade_result := upgrade_procedure()
                return_message += upgrade_result
        }
	end := time.Now()
	elapsed_time := end.Sub(start)

	result += "\nJuju Zero Impact Upgrade - " + os.Args[1]
	result += return_message
	result += "\nStarted at " + start.String()
	result += "\nEnded at " + end.String()
	result += "\nTime taken = " + elapsed_time.String()
}

func write_result() {
	file := "result.txt"
        remove_previous(file)
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
