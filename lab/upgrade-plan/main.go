package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var debug = os.Getenv("DEBUG") != ""
var counter = 0

type DisruptionBudget struct {
	AppName           string
	DisruptionAllowed int
}

type Node struct {
	NodeName string
}

// Application represents an instance, like a Pod
type Application struct {
	AppName  string
	NodeName string
}

type Calculator struct {
	pods map[string]map[string]bool
	memo map[[16]byte][]string
}

// upgradeStep finds most nodes that can be upgraded at once:
//
// this can be solved by dynamic programming because it meets two requirements:
// - optimal substructure: whether we choose to calculate a node or not,
//   the optimal solution of the problem is also the optimal solution of the rest nodes;
// - overlapping sub-problems: a node may not affect another if they don't have same apps running,
//   so the result can be reused;
//
// transition equation:
//
// S(node, budgets) = Max of:
//       a. not calculate n1: S(node(without n1), budgets)
//       b. calculate n1:     1 + S(node(without n1), budgets(minus pods on n1 because n1 will be upgraded))
//
func (c *Calculator) upgradeStep(nodes []string, budgets map[string]int) (steps []string) {
	if debug {
		counter++
		log.Println(counter, nodes, budgets)
	}

	if len(nodes) == 0 {
		return nil
	}

	stepWhenNotUpgradeFirstNode := c.upgradeStep(nodes[1:], budgets)

	canUpgradeFirstNode := true
	appsOnFirstNode := c.pods[nodes[0]]
	budgetsIfUpgradeFirstNode := make(map[string]int)
	for app := range budgets {
		if appsOnFirstNode[app] {
			budgetsIfUpgradeFirstNode[app] = budgets[app] - 1
			if budgetsIfUpgradeFirstNode[app] < 0 {
				canUpgradeFirstNode = false
				break
			}
		} else {
			budgetsIfUpgradeFirstNode[app] = budgets[app]
		}
	}

	if !canUpgradeFirstNode {
		steps = stepWhenNotUpgradeFirstNode
	} else {
		stepWhenUpgradeFirstNode := c.upgradeStep(nodes[1:], budgetsIfUpgradeFirstNode)
		if len(stepWhenUpgradeFirstNode)+1 > len(stepWhenNotUpgradeFirstNode) {
			steps = append([]string{nodes[0]}, stepWhenUpgradeFirstNode...)
		} else {
			steps = stepWhenNotUpgradeFirstNode
		}
	}
	return steps
}

// calculate generates an upgrade plan
func (c *Calculator) calculate(nodes []string, budgets map[string]int) [][]string {
	log.Println("calculating...")
	var plan [][]string
	for len(nodes) > 0 {
		var nodesLeft []string
		step := c.upgradeStep(nodes, budgets)
		log.Printf("step calculated: %v", step)
		for _, node := range nodes {
			if !stringInSlice(node, step) {
				nodesLeft = append(nodesLeft, node)
			}
		}
		plan = append(plan, step)
		nodes = nodesLeft
	}
	return plan
}

// Generate generates an upgrade plan
func (c *Calculator) Generate(nodes []Node, pods []Application, budgets []DisruptionBudget) [][]string {
	log.Println("preparing...")
	var nodeNames []string
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.NodeName)
	}
	podsOnNode := make(map[string]map[string]bool)
	for _, pod := range pods {
		if podsOnNode[pod.NodeName] == nil {
			podsOnNode[pod.NodeName] = make(map[string]bool)
		}
		podsOnNode[pod.NodeName][pod.AppName] = true
	}
	budgetMap := make(map[string]int)
	for _, budget := range budgets {
		budgetMap[budget.AppName] = budget.DisruptionAllowed
	}
	c.pods = podsOnNode
	return c.calculate(nodeNames, budgetMap)
}

func stringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

type Testcase struct {
	Nodes   []Node
	Pods    []Application
	Budgets []DisruptionBudget
}

var testcases = []Testcase{
	{
		Nodes: []Node{
			{NodeName: "n1"},
			{NodeName: "n2"},
			{NodeName: "n3"},
		},
		Pods: []Application{
			{AppName: "app1", NodeName: "n1"},
			{AppName: "app1", NodeName: "n2"},
			{AppName: "app2", NodeName: "n1"},
			{AppName: "app2", NodeName: "n2"},
			{AppName: "app3", NodeName: "n2"},
			{AppName: "app3", NodeName: "n3"},
		},
		Budgets: []DisruptionBudget{
			{AppName: "app1", DisruptionAllowed: 1},
			{AppName: "app2", DisruptionAllowed: 1},
			{AppName: "app3", DisruptionAllowed: 1},
		},
	},
}

// main
//
// To be considered:
// - pods may be changed (scaled, rescheduled) during operation;
// - apps may have more complex disruption restrictions;
func main() {
	rand.Seed(time.Now().Unix())

	if len(os.Args) < 3 {
		fmt.Println("usage: go run . [action]\n" +
			"\n" +
			"go run . testcase 0 # test specific testcase, index: 0\n" +
			"go run . random 10 5 # test random generated testcase, 10 nodes, 5 apps")
		return
	}

	action := os.Args[1]

	var testcase Testcase
	switch action {
	case "testcase":
		// test specific testcase
		if len(os.Args) < 3 {
			fmt.Println("arg missing")
			return
		}

		n, _ := strconv.Atoi(os.Args[2])

		if n < 0 || n > len(testcases)-1 {
			fmt.Printf("undefined testcase: %s\n", os.Args[2])
			return
		}
		testcase = testcases[n]
	case "random":
		// test random generated testcase
		if len(os.Args) < 4 {
			fmt.Println("arg missing")
			return
		}

		nNodes, _ := strconv.Atoi(os.Args[2])
		nApps, _ := strconv.Atoi(os.Args[3])

		fmt.Println("generating random testcase...")
		var nodes []Node
		var pods []Application
		var budgets []DisruptionBudget
		for i := 0; i < nNodes; i++ {
			nodes = append(nodes, Node{NodeName: fmt.Sprintf("n%d", i+1)})
		}
		for i := 0; i < nApps; i++ {
			var expectNumberOfPods int
			if nNodes < 200 {
				expectNumberOfPods = rand.Intn(nNodes) // like 2/3, 3/5
			} else {
				expectNumberOfPods = rand.Intn(200) // like 60/200, 80/5000, not too many
			}
			actualNumberOfPods := 0
			for j := 0; j < nNodes; j++ {
				if rand.Intn(nNodes) < expectNumberOfPods {
					pods = append(pods, Application{
						AppName:  fmt.Sprintf("app%d", i+1),
						NodeName: fmt.Sprintf("n%d", j+1),
					})
					actualNumberOfPods += 1
				}
			}
			if actualNumberOfPods > 0 {
				var disruptionAllowed int
				if actualNumberOfPods < 100 {
					disruptionAllowed = rand.Intn(actualNumberOfPods) + 1
				} else {
					disruptionAllowed = rand.Intn(100) + 1
				}
				budgets = append(budgets, DisruptionBudget{
					AppName: fmt.Sprintf("app%d", i+1), DisruptionAllowed: disruptionAllowed})
			}
		}
		testcase = Testcase{
			Nodes:   nodes,
			Pods:    pods,
			Budgets: budgets,
		}
	default:
		fmt.Printf("unknown action: %s\n", action)
		return
	}

	calculator := Calculator{
		memo: make(map[[16]byte][]string),
	}

	fmt.Printf("\nnodes:\n")
	for _, node := range testcase.Nodes {
		var podsOnNode []string
		for _, pod := range testcase.Pods {
			if pod.NodeName == node.NodeName {
				podsOnNode = append(podsOnNode, pod.AppName)
			}
		}
		fmt.Printf("  %s: %v\n", node.NodeName, podsOnNode)
	}
	fmt.Println("budgets:")
	for _, budget := range testcase.Budgets {
		fmt.Printf("  %s: %d\n", budget.AppName, budget.DisruptionAllowed)
	}
	fmt.Println()

	start := time.Now()
	_ = calculator.Generate(testcase.Nodes, testcase.Pods, testcase.Budgets)
	end := time.Now()

	fmt.Printf("\ntime spent: %v\n", end.Sub(start))
}
