package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type JobData struct {
	Color      string
	Jobid      int
	Location   string
	Locationf  string
	Mode       string
	Nodes      int
	Project    string
	Queue      string
	Runtime    int
	Runtimef   string
	Starttime  string
	State      string
	Submittime float32
	Walltime   int
	Walltimef  string
}

type JobDatas []*JobData

func (s JobDatas) Len() int      { return len(s) }
func (s JobDatas) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type JobDataByWalltime struct {
	JobDatas
}

func (s JobDataByWalltime) Less(i, j int) bool { return s.JobDatas[i].Walltime > s.JobDatas[j].Walltime }

type ReservationData struct {
	Duration   int
	Durationf  string
	Name       string
	Partitions string
	Queue      string
	Start      float32
	Startf     string
	Tminus     string
}

type DimensionData struct {
	Midplanes    int
	Nodecards    int
	Racks        int
	Rows         int
	Subdivisions int
}

type GronkData struct {
	Updated     int64
	Dimensions  DimensionData
	Running     JobDatas
	Queued      JobDatas
	Reservation []ReservationData
}

func printHeader(machine string) {
	fmt.Println(strings.Repeat("-", 108))
	fmt.Printf("|    %-101s |\n", machine+" job data")
	fmt.Println(strings.Repeat("-", 108))
	fmt.Printf("| %-7s| %-16s| %-9s| %-9s| %-20s| %-10s| %-8s| %-12s|\n",
		"Job Id", "Project", "Run Time", "Walltime", "Location",
		"Queue", "Nodes", "Mode")
	fmt.Println(strings.Repeat("-", 108))
}

func printFooter(updated int64) {
	fmt.Println(strings.Repeat("-", 108))
	fmt.Printf("| Last updated: %-90s |\n", time.Unix(updated, 0))
	fmt.Println(strings.Repeat("-", 108))
}

func printJobData(machine string, gronkdata GronkData) {
	fmt.Print("\033[2J") // clear the screen
	printHeader(machine)
	sort.Sort(JobDataByWalltime{gronkdata.Running})
	//TODO: Handle empty running jobs.
	for _, job := range gronkdata.Running {
		fmt.Printf("| %-7d| %-16s| %-9s| %-9s| %-20s| %-10s| %-8d| %-12s|\n",
			job.Jobid, job.Project, job.Runtimef, job.Walltimef, job.Locationf,
			job.Queue, job.Nodes, job.Mode)
	}
	printFooter(gronkdata.Updated)
}

func getJobData(machine string) (GronkData, error) {
	var gronkdata GronkData
	url := fmt.Sprintf("http://status.alcf.anl.gov/%s/activity.json", machine)

	r, err := http.Get(url)
	if err != nil {
		return gronkdata, err
	}

	dec := json.NewDecoder(r.Body)
	dec.Decode(&gronkdata)
	r.Body.Close()

	return gronkdata, nil
}

func main() {
	//TODO: better arg parsing, handle no args
	machine := os.Args[1]
	if machine == "" {
		fmt.Println("TYPE IN A MACHINE DUMMY!")
		os.Exit(1)
	}
	dataChan := make(chan GronkData)

	go func() {
		for {
			data, err := getJobData(machine)
			if err != nil {
				fmt.Println("Oops!:", err)
				os.Exit(1)
			}
			dataChan <- data
			time.Sleep(time.Second * 5)
		}
	}()

	for data := range dataChan {
		printJobData(machine, data)
	}
}
