package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/moby/moby/client"
	"github.com/spf13/cobra"
)

// healthcheckCmd represents the healthcheck command
var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Check the health of Ghostwriter's services",
	Long: `Check the health of Ghostwriter's services.

This command validates all containers are running and passing
their respective health checks`,
	Run: runHealthcheck,
}

func init() {
	rootCmd.AddCommand(healthcheckCmd)
}

func runHealthcheck(cmd *cobra.Command, args []string) {
	dockerInterface := docker.GetDockerInterface(mode)
	// initialize tabwriter
	writer := new(tabwriter.Writer)
	// Set minwidth, tabwidth, padding, padchar, and flags
	writer.Init(os.Stdout, 8, 8, 1, '\t', 0)

	defer writer.Flush()

	fmt.Println("[+] Checking Ghostwriter containers and their respective health checks...")

	containerIssues, dockerErr := checkDockerHealth(dockerInterface)

	if dockerErr != nil {
		fmt.Printf("[!] Failed to get container information from Docker: %s\n", dockerErr)
	} else {
		if len(containerIssues) > 0 {
			fmt.Printf("[*] Identified %d issues with one or more containers:\n\n", len(containerIssues))

			fmt.Fprintf(writer, "\n %s\t%s\t%s", "Type", "Container", "Message")
			fmt.Fprintf(writer, "\n %s\t%s\t%s", "––––––––––––", "––––––––––––", "––––––––––––")

			for _, issue := range containerIssues {
				fmt.Fprintf(writer, "\n %s\t%s\t%s", issue.Type, issue.Service, issue.Message)
			}
		} else {
			fmt.Println("[*] Identified zero container issues, now testing services...")
			serviceIssues, svcErr := checkGhostwriterHealth(dockerInterface)
			if svcErr != nil {
				fmt.Printf("[!] Failed to get health status from Ghostwriter's /status/ endpoint: %s\n", svcErr)
			} else {
				if len(serviceIssues) > 0 {
					fmt.Printf("[*] Identified %d issues with one or more services:\n\n", len(serviceIssues))

					fmt.Fprintf(writer, "\n %s\t%s\t%s", "Type", "Service", "Message")
					fmt.Fprintf(writer, "\n %s\t%s\t%s", "––––––––––––", "––––––––––––", "––––––––––––")

					for _, issue := range serviceIssues {
						fmt.Fprintf(writer, "\n %s\t%s\t%s", issue.Type, issue.Service, issue.Message)
					}
				} else {
					fmt.Println("[*] Identified zero issues with core services")
				}
			}
		}
	}
}

type HealthIssue struct {
	Type    string
	Service string
	Message string
}

type HealthIssues []HealthIssue

func (c HealthIssues) Len() int {
	return len(c)
}

func (c HealthIssues) Less(i, j int) bool {
	return c[i].Service < c[j].Service
}

func (c HealthIssues) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func checkDockerHealth(dockerInterface *docker.DockerInterface) (HealthIssues, error) {
	var found []string
	var imageName string
	var issues HealthIssues

	requiredImages := docker.ProdImages
	if dockerInterface.UseDevInfra {
		requiredImages = docker.DevImages
	}

	// Check running containers to make sure every necessary container is up
	cli, err := dockerInterface.GetDaemonClient()
	if err != nil {
		return issues, err
	}

	containers, err := cli.ContainerList(context.Background(), client.ContainerListOptions{
		All: false,
	})
	if err != nil {
		return issues, err
	}

	if len(containers.Items) > 0 {
		for _, container := range containers.Items {
			if docker.Contains(docker.DevImages, container.Image) || docker.Contains(docker.ProdImages, container.Image) {
				found = append(found, container.Image)
			}
		}
		for _, image := range requiredImages {
			if !docker.Contains(found, image) {
				imageName = strings.ToUpper(image[strings.LastIndex(image, "_")+1:])
				issues = append(issues, HealthIssue{"Container", imageName, "Container is not running"})
			}
		}
	} else {
		issues = append(issues, HealthIssue{"Container", "ALL", "No Ghostwriter containers are running"})
	}

	return issues, nil
}

// CheckGhostwriterHealth fetches the latest health reports from Ghostwriter's status API endpoint.
func checkGhostwriterHealth(dockerInterface *docker.DockerInterface) (HealthIssues, error) {
	var issues HealthIssues

	protocol := "https"
	port := "443"
	if dockerInterface.UseDevInfra {
		protocol = "http"
		port = "8000"
	}

	baseUrl := protocol + "://localhost:" + port + "/status/"
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := http.Client{Timeout: time.Second * 2, Transport: transport}

	req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
	if err != nil {
		return issues, err
	}

	req.Header.Set("Accept", "application/json")

	res, getErr := client.Do(req)
	if getErr != nil {
		return issues, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return issues, errors.New("Non-OK HTTP status suggests an issue with the Django or Nginx services (Code " + strconv.Itoa(res.StatusCode) + ")")
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return issues, readErr
	}

	var results map[string]interface{}
	jsonErr := json.Unmarshal(body, &results)
	if jsonErr != nil {
		return issues, jsonErr
	}

	for key := range results {
		if results[key] != "working" {
			var statusMsg string
			if str, ok := results[key].(string); ok {
				statusMsg = str
			} else {
				statusMsg = fmt.Sprint(results[key])
			}
			issues = append(issues, HealthIssue{"Service", key, statusMsg})
		}
	}

	return issues, nil
}
