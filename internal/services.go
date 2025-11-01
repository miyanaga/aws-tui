package internal

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bporter816/aws-tui/internal/model"
	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/settings"
	"github.com/bporter816/aws-tui/internal/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

//go:generate go run gen.go arg1

type Services struct {
	*ui.Tree
	repos          map[string]interface{}
	app            *Application
	searchBuffer   string
	allNodes       []*tview.TreeNode
	searchTimer    *time.Timer
	settings       *settings.Settings
	favoritesNode  *tview.TreeNode
	servicesNode   *tview.TreeNode
	serviceMap     map[string][]string
}

func NewServices(repos map[string]interface{}, app *Application) *Services {
	m := map[string][]string{
		"ACM": {
			"Certificates",
		},
		"ACM PCA": {
			"Certificate Authorities",
		},
		"CloudFront": {
			"Distributions",
			"Functions",
		},
		"CloudWatch": {
			"Log Groups",
		},
		"DynamoDB": {
			"Tables",
		},
		"EBS": {
			"Volumes",
		},
		"EC2": {
			"Instances",
			"Availability Zones",
			"Security Groups",
			"AMIs",
			"Key Pairs",
			"Reserved Instances",
		},
		"ECS": {
			"Clusters",
			"Task Definitions",
		},
		"EKS": {
			"Clusters",
		},
		"ELB": {
			"Load Balancers",
			"Target Groups",
			"Trust Stores",
		},
		"ElastiCache": {
			"Clusters",
			"Users",
			"Groups",
			"Parameter Groups",
			"Subnet Groups",
			"Reserved Nodes",
			"Snapshots",
			"Events",
			"Service Updates",
		},
		"Global Accelerator": {
			"Accelerators",
		},
		"IAM": {
			"Users",
			"Roles",
			"Groups",
			"Managed Policies",
		},
		"KMS": {
			"Keys",
			"Custom Key Stores",
		},
		"Lambda": {
			"Functions",
		},
		"MQ": {
			"Brokers",
		},
		"MSK": {
			"Clusters",
		},
		"RDS": {
			"Clusters",
			"Global Clusters",
			"Parameter Groups",
			"Subnet Groups",
			"Reserved Instances",
		},
		"Route 53": {
			"Hosted Zones",
			"Health Checks",
		},
		"S3": {
			"Buckets",
		},
		"SNS": {
			"Topics",
		},
		"SQS": {
			"Queues",
		},
		"Secrets Manager": {
			"Secrets",
		},
		"Service Quotas": {
			"Services",
		},
		"Systems Manager": {
			"Parameters",
		},
		"VPC": {
			"VPCs",
			"Subnets",
			"Internet Gateways",
		},
	}

	// Load settings
	userSettings, err := settings.Load()
	if err != nil {
		userSettings = &settings.Settings{Favorites: []string{}}
	}

	root := tview.NewTreeNode("")
	s := &Services{
		Tree:         ui.NewTree(root),
		repos:        repos,
		app:          app,
		searchBuffer: "",
		allNodes:     make([]*tview.TreeNode, 0),
		settings:     userSettings,
		serviceMap:   m,
	}

	s.buildTree()
	s.SetSelectedFunc(s.selectHandler)
	s.setupSearchCapture()

	// Set default focus to first favorite if exists
	if len(userSettings.Favorites) > 0 && s.favoritesNode != nil {
		s.favoritesNode.Expand()
		children := s.favoritesNode.GetChildren()
		if len(children) > 0 {
			s.SetCurrentNode(children[0])
		}
	}

	return s
}

func (s *Services) buildTree() {
	// Clear existing tree
	s.Root.ClearChildren()
	s.allNodes = make([]*tview.TreeNode, 0)

	// Build Favorites section if favorites exist
	if len(s.settings.Favorites) > 0 {
		s.favoritesNode = tview.NewTreeNode("Favorites")
		s.Root.AddChild(s.favoritesNode)

		for _, fav := range s.settings.Favorites {
			parts := strings.SplitN(fav, ".", 2)
			if len(parts) == 2 {
				displayText := parts[0] + " > " + parts[1]
				leaf := tview.NewTreeNode(displayText)
				leaf.SetReference(fav)
				s.favoritesNode.AddChild(leaf)
				s.allNodes = append(s.allNodes, leaf)
			}
		}
		s.favoritesNode.Expand()
	}

	// Build Services section
	s.servicesNode = tview.NewTreeNode("Services")
	s.Root.AddChild(s.servicesNode)

	// Sort service keys
	var keys []string
	for k := range s.serviceMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := s.serviceMap[k]
		n := tview.NewTreeNode(k)
		s.servicesNode.AddChild(n)
		s.allNodes = append(s.allNodes, n)
		for _, view := range v {
			leaf := tview.NewTreeNode(view)
			leaf.SetReference(fmt.Sprintf("%v.%v", k, view))
			n.AddChild(leaf)
			s.allNodes = append(s.allNodes, leaf)
		}
		n.CollapseAll()
	}
	s.servicesNode.Expand()
}

func (s Services) GetService() string {
	return "Services"
}

func (s Services) GetLabels() []string {
	return []string{}
}

func (s Services) selectHandler(n *tview.TreeNode) {
	// Handle root nodes (Favorites, Services) and service category nodes
	ref := n.GetReference()
	if ref == nil {
		// This is a Favorites or Services node, or a service category
		if n.IsExpanded() {
			n.Collapse()
		} else {
			n.Expand()
			if len(n.GetChildren()) > 0 {
				s.SetCurrentNode(n.GetChildren()[0])
			}
		}
		return
	}

	view := ref.(string)
	var item Component
	switch view {
	case "ACM.Certificates":
		item = NewACMCertificates(s.repos["ACM"].(*repo.ACM), s.app)
	case "ACM PCA.Certificate Authorities":
		item = NewACMPCACertificateAuthorities(s.repos["ACM PCA"].(*repo.ACMPCA), s.app)
	case "CloudFront.Distributions":
		item = NewCFDistributions(s.repos["CloudFront"].(*repo.CloudFront), s.app)
	case "CloudFront.Functions":
		item = NewCFFunctions(s.repos["CloudFront"].(*repo.CloudFront), s.app)
	case "CloudWatch.Log Groups":
		item = NewCloudWatchLogGroups(s.repos["CloudWatch"].(*repo.CloudWatch), s.app)
	case "DynamoDB.Tables":
		item = NewDynamoDBTables(s.repos["DynamoDB"].(*repo.DynamoDB), s.app)
	case "EBS.Volumes":
		item = NewEBSVolumes(s.repos["EC2"].(*repo.EC2), s.app)
	case "EC2.Instances":
		item = NewEC2Instances(s.repos["EC2"].(*repo.EC2), s.app)
	case "EC2.Availability Zones":
		item = NewEC2AvailabilityZones(s.repos["EC2"].(*repo.EC2), s.app)
	case "EC2.Security Groups":
		item = NewEC2SecurityGroups(s.repos["EC2"].(*repo.EC2), s.app)
	case "EC2.AMIs":
		item = NewEC2Images(s.repos["EC2"].(*repo.EC2), s.app)
	case "EC2.Key Pairs":
		item = NewEC2KeyPairs(s.repos["EC2"].(*repo.EC2), s.app)
	case "EC2.Reserved Instances":
		item = NewEC2ReservedInstances(s.repos["EC2"].(*repo.EC2), s.app)
	case "ECS.Clusters":
		item = NewECSClusters(s.repos["ECS"].(*repo.ECS), s.app)
	case "ECS.Task Definitions":
		item = NewECSTaskDefinitions(s.repos["ECS"].(*repo.ECS), s.app)
	case "EKS.Clusters":
		item = NewEKSClusters(s.repos["EKS"].(*repo.EKS), s.app)
	case "ELB.Load Balancers":
		item = NewELBLoadBalancers(s.repos["ELB"].(*repo.ELB), s.app)
	case "ELB.Target Groups":
		item = NewELBTargetGroups(s.repos["ELB"].(*repo.ELB), s.app)
	case "ELB.Trust Stores":
		item = NewELBTrustStores(s.repos["ELB"].(*repo.ELB), s.app)
	case "ElastiCache.Clusters":
		item = NewElastiCacheClusters(s.repos["ElastiCache"].(*repo.ElastiCache), s.app)
	case "ElastiCache.Users":
		item = NewElastiCacheUsers(s.repos["ElastiCache"].(*repo.ElastiCache), s.app)
	case "ElastiCache.Groups":
		item = NewElastiCacheGroups(s.repos["ElastiCache"].(*repo.ElastiCache), s.app)
	case "ElastiCache.Parameter Groups":
		item = NewElastiCacheParameterGroups(s.repos["ElastiCache"].(*repo.ElastiCache), s.app)
	case "ElastiCache.Subnet Groups":
		item = NewElastiCacheSubnetGroups(s.repos["ElastiCache"].(*repo.ElastiCache), s.repos["EC2"].(*repo.EC2), s.app)
	case "ElastiCache.Reserved Nodes":
		item = NewElastiCacheReservedCacheNodes(s.repos["ElastiCache"].(*repo.ElastiCache), s.app)
	case "ElastiCache.Snapshots":
		item = NewElastiCacheSnapshots(s.repos["ElastiCache"].(*repo.ElastiCache), s.app)
	case "ElastiCache.Events":
		item = NewElastiCacheEvents(s.repos["ElastiCache"].(*repo.ElastiCache), s.app)
	case "ElastiCache.Service Updates":
		item = NewElastiCacheServiceUpdates(s.repos["ElastiCache"].(*repo.ElastiCache), s.app)
	case "Global Accelerator.Accelerators":
		item = NewGlobalAcceleratorAccelerators(s.repos["Global Accelerator"].(*repo.GlobalAccelerator), s.app)
	case "IAM.Users":
		item = NewIAMUsers(s.repos["IAM"].(*repo.IAM), nil, s.app)
	case "IAM.Roles":
		item = NewIAMRoles(s.repos["IAM"].(*repo.IAM), s.app)
	case "IAM.Groups":
		item = NewIAMGroups(s.repos["IAM"].(*repo.IAM), nil, s.app)
	case "IAM.Managed Policies":
		item = NewIAMPolicies(s.repos["IAM"].(*repo.IAM), model.IAMIdentityTypeAll, nil, s.app)
	case "KMS.Keys":
		item = NewKmsKeys(s.repos["KMS"].(*repo.KMS), s.app)
	case "KMS.Custom Key Stores":
		item = NewKmsCustomKeyStores(s.repos["KMS"].(*repo.KMS), s.app)
	case "Lambda.Functions":
		item = NewLambdaFunctions(s.repos["Lambda"].(*repo.Lambda), s.app)
	case "MQ.Brokers":
		item = NewMQBrokers(s.repos["MQ"].(*repo.MQ), s.app)
	case "MSK.Clusters":
		item = NewMSKClusters(s.repos["MSK"].(*repo.MSK), s.app)
	case "RDS.Clusters":
		item = NewRDSClusters(s.repos["RDS"].(*repo.RDS), s.app)
	case "RDS.Global Clusters":
		item = NewRDSGlobalClusters(s.repos["RDS"].(*repo.RDS), s.app)
	case "RDS.Parameter Groups":
		item = NewRDSParameterGroups(s.repos["RDS"].(*repo.RDS), s.app)
	case "RDS.Subnet Groups":
		item = NewRDSSubnetGroups(s.repos["RDS"].(*repo.RDS), s.repos["EC2"].(*repo.EC2), s.app)
	case "RDS.Reserved Instances":
		item = NewRDSReservedInstances(s.repos["RDS"].(*repo.RDS), s.app)
	case "Route 53.Hosted Zones":
		item = NewRoute53HostedZones(s.repos["Route 53"].(*repo.Route53), s.app)
	case "Route 53.Health Checks":
		item = NewRoute53HealthChecks(s.repos["Route 53"].(*repo.Route53), s.app)
	case "S3.Buckets":
		item = NewS3Buckets(s.repos["S3"].(*repo.S3), s.app)
	case "SNS.Topics":
		item = NewSNSTopics(s.repos["SNS"].(*repo.SNS), s.app)
	case "SQS.Queues":
		item = NewSQSQueues(s.repos["SQS"].(*repo.SQS), s.app)
	case "Secrets Manager.Secrets":
		item = NewSMSecrets(s.repos["Secrets Manager"].(*repo.SecretsManager), s.app)
	case "Service Quotas.Services":
		item = NewServiceQuotasServices(s.repos["Service Quotas"].(*repo.ServiceQuotas), s.app)
	case "Systems Manager.Parameters":
		item = NewSSMParameters(s.repos["SSM"].(*repo.SSM), s.app)
	case "VPC.VPCs":
		item = NewVPCVPCs(s.repos["EC2"].(*repo.EC2), s.app)
	case "VPC.Subnets":
		item = NewVPCSubnets(s.repos["EC2"].(*repo.EC2), []string{}, "", s.app)
	case "VPC.Internet Gateways":
		item = NewVPCInternetGateways(s.repos["EC2"].(*repo.EC2), s.app)
	default:
		panic("unknown service")
	}
	s.app.AddAndSwitch(item)
}

func (s *Services) setupSearchCapture() {
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle 'd' key to add to favorites
		if event.Key() == tcell.KeyRune && event.Rune() == 'd' {
			s.addToFavorites()
			return nil
		}

		// Handle 'x' key to remove from favorites
		if event.Key() == tcell.KeyRune && event.Rune() == 'x' {
			s.removeFromFavorites()
			return nil
		}

		// Handle backspace in search
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			if len(s.searchBuffer) > 0 {
				s.searchBuffer = s.searchBuffer[:len(s.searchBuffer)-1]
				s.updateSearch()
				s.resetSearchTimer()
				return nil
			}
			return event
		}

		// Handle printable characters for search
		if event.Key() == tcell.KeyRune {
			s.searchBuffer += string(event.Rune())
			s.updateSearch()
			s.resetSearchTimer()
			return nil
		}

		// Clear search on Escape
		if event.Key() == tcell.KeyEscape {
			s.clearSearch()
			return event
		}

		// Other keys (navigation, enter, etc.) clear search
		if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			s.clearSearch()
			return event
		}

		return event
	})
}

func (s *Services) addToFavorites() {
	node := s.GetCurrentNode()
	if node == nil {
		return
	}

	ref := node.GetReference()
	if ref == nil {
		return
	}

	view := ref.(string)
	if err := s.settings.AddFavorite(view); err == nil {
		// Rebuild tree to show new favorite
		currentRef := view
		s.buildTree()
		// Try to restore focus to the added item in Favorites
		if s.favoritesNode != nil {
			for _, child := range s.favoritesNode.GetChildren() {
				if child.GetReference() == currentRef {
					s.SetCurrentNode(child)
					return
				}
			}
		}
	}
}

func (s *Services) removeFromFavorites() {
	node := s.GetCurrentNode()
	if node == nil {
		return
	}

	ref := node.GetReference()
	if ref == nil {
		return
	}

	// Check if we're in the Favorites section
	parent := s.findParent(node)
	if parent != s.favoritesNode {
		return
	}

	view := ref.(string)
	if err := s.settings.RemoveFavorite(view); err == nil {
		// Rebuild tree to remove favorite
		s.buildTree()
		// Set focus to Favorites node or first service
		if s.favoritesNode != nil && len(s.favoritesNode.GetChildren()) > 0 {
			s.SetCurrentNode(s.favoritesNode.GetChildren()[0])
		} else if s.servicesNode != nil {
			s.SetCurrentNode(s.servicesNode)
		}
	}
}

func (s *Services) resetSearchTimer() {
	// Stop existing timer if any
	if s.searchTimer != nil {
		s.searchTimer.Stop()
	}

	// Create new timer that will clear search after 1 second
	s.searchTimer = time.AfterFunc(1*time.Second, func() {
		s.app.app.QueueUpdateDraw(func() {
			s.clearSearch()
		})
	})
}

func (s *Services) clearSearch() {
	s.searchBuffer = ""
	s.SetTitle("")
	if s.searchTimer != nil {
		s.searchTimer.Stop()
		s.searchTimer = nil
	}
}

func (s *Services) updateSearch() {
	if s.searchBuffer == "" {
		s.clearSearch()
		return
	}

	// Show search buffer in title
	s.SetTitle(fmt.Sprintf(" Search: %s ", s.searchBuffer))

	// Find matching node
	searchLower := strings.ToLower(s.searchBuffer)
	for _, node := range s.allNodes {
		nodeText := strings.ToLower(node.GetText())
		if strings.HasPrefix(nodeText, searchLower) {
			// Expand parent if this is a child node
			parent := s.findParent(node)
			if parent != nil {
				parent.Expand()
			}
			s.SetCurrentNode(node)
			return
		}
	}
}

func (s *Services) findParent(target *tview.TreeNode) *tview.TreeNode {
	var findParentRecursive func(node *tview.TreeNode) *tview.TreeNode
	findParentRecursive = func(node *tview.TreeNode) *tview.TreeNode {
		for _, child := range node.GetChildren() {
			if child == target {
				return node
			}
			if found := findParentRecursive(child); found != nil {
				return found
			}
		}
		return nil
	}
	return findParentRecursive(s.Root)
}

func (s Services) GetKeyActions() []KeyAction {
	return []KeyAction{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
			Description: "Add to Favorites",
			Action:      s.addToFavorites,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone),
			Description: "Remove from Favorites",
			Action:      s.removeFromFavorites,
		},
	}
}

func (s Services) Render() {
}
