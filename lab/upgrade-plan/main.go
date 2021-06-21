package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var debug = os.Getenv("DEBUG") != ""

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

// upgradeStep finds most nodes that can be upgraded at once:
//
// this can be solved by dynamic programming because it meets two requirements:
// - optimal substructure: whether we choose to upgrade a node or not,
//   the optimal solution of the problem is also the optimal solution of the rest nodes;
// - overlapping sub-problems: a node may not affect another if they don't have same apps running,
//   so the result can be reused;
//
// transition equation:
//
// S(node, budgets) = Max of:
//       a. not upgrade n1: S(node(without n1), budgets)
//       b. upgrade n1:     1 + S(node(without n1), budgets(minus pods on n1 because n1 will be upgraded))
//
func upgradeStep(indent string, nodes []string, pods map[string][]string, budgets map[string]int) []string {
	if debug {
		fmt.Println(indent, nodes, budgets)
	}

	if len(nodes) == 0 {
		return nil
	}

	stepWhenNotUpgradeFirstNode := upgradeStep(indent+" -", nodes[1:], pods, budgets)

	canUpgradeFirstNode := true
	appsOnFirstNode := pods[nodes[0]]
	budgetsIfUpgradeFirstNode := make(map[string]int)
	for app := range budgets {
		if stringInSlice(app, appsOnFirstNode) {
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
		return stepWhenNotUpgradeFirstNode
	}

	stepWhenUpgradeFirstNode := upgradeStep(indent+" +", nodes[1:], pods, budgetsIfUpgradeFirstNode)
	if len(stepWhenUpgradeFirstNode)+1 > len(stepWhenNotUpgradeFirstNode) {
		return append([]string{nodes[0]}, stepWhenUpgradeFirstNode...)
	}
	return stepWhenNotUpgradeFirstNode
}

// upgrade generates an upgrade plan
func upgrade(nodes []string, pods map[string][]string, budgets map[string]int) [][]string {
	var plan [][]string
	for len(nodes) > 0 {
		var nodesLeft []string
		step := upgradeStep("", nodes, pods, budgets)
		if debug {
			fmt.Printf("step calculated: %v\n\n", step)
		}
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

// Upgrade generates an upgrade plan
func Upgrade(nodes []Node, pods []Application, budgets []DisruptionBudget) [][]string {
	var nodeNames []string
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.NodeName)
	}
	podsOnNode := make(map[string][]string)
	for _, pod := range pods {
		podsOnNode[pod.NodeName] = append(podsOnNode[pod.NodeName], pod.AppName)
	}
	budgetMap := make(map[string]int)
	for _, budget := range budgets {
		budgetMap[budget.AppName] = budget.DisruptionAllowed
	}
	return upgrade(nodeNames, podsOnNode, budgetMap)
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
			"go run . test 0 # test specific case, index: 0\n" +
			"go run . random 10 5 # test random generated case, 10 nodes, 5 apps")
		return
	}

	action := os.Args[1]

	var testcase Testcase
	switch action {
	case "test":
		// test specific case
		if len(os.Args) < 3 {
			fmt.Printf("arg missing")
			return
		}

		n, _ := strconv.Atoi(os.Args[2])

		if n < 0 || n > len(testcases)-1 {
			fmt.Printf("undefined testcase: %s\n", os.Args[2])
			return
		}
		testcase = testcases[n]
	case "random":
		// test random generated case
		if len(os.Args) < 4 {
			fmt.Printf("arg missing")
			return
		}

		nNodes, _ := strconv.Atoi(os.Args[2])
		nApps, _ := strconv.Atoi(os.Args[3])

		var nodes []Node
		var pods []Application
		var budgets []DisruptionBudget
		for i := 0; i < nNodes; i++ {
			nodes = append(nodes, Node{NodeName: fmt.Sprintf("n%d", i+1)})
		}
		for i := 0; i < nApps; i++ {
			expectNumberOfPods := rand.Intn(nNodes)
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
				budgets = append(budgets, DisruptionBudget{
					AppName: fmt.Sprintf("app%d", i+1), DisruptionAllowed: rand.Intn(actualNumberOfPods) + 1})
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

	start := time.Now()
	plan := Upgrade(testcase.Nodes, testcase.Pods, testcase.Budgets)
	end := time.Now()

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
	fmt.Printf("budgets:\n")
	for _, budget := range testcase.Budgets {
		fmt.Printf("  %s: %d\n", budget.AppName, budget.DisruptionAllowed)
	}
	fmt.Printf("plan:\n")
	for j, step := range plan {
		fmt.Printf("  step %d: %v\n", j, step)
	}
	fmt.Printf("time spent: %v\n", end.Sub(start))
}
