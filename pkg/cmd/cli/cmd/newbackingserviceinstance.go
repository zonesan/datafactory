package cmd

import (
	"errors"
	"fmt"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
	"github.com/spf13/cobra"
	"io"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	newBackingServiceInstanceLong = `
Create a new BackingServiceInstance

`
	newBackingServiceInstanceExample = `# Create a new backingserviceinstance with [name BackingServiceName BackingServicePlanGuid DashboardURL]
  $ %[1]s  mysql_BackingServiceInstance --backingservice_name="BackingServiceName" --backingservice_plan_guid="BackingServicePlanGuid" --dashboard_url="DashboardUrl"`
)

type NewBackingServiceInstanceOptions struct {
	Name      string
	
	DashboardUrl           string
	BackingServiceName     string
	BackingServicePlanGuid string

	Client client.Interface

	Out io.Writer
}

func NewCmdNewBackingServiceInstance(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &NewBackingServiceInstanceOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     "new-backingserviceinstance NAME --backingservice_name=BackingServiceName --plan_guid=BackingServicePlanGuid --dashboard_url=DashboardUrl",
		Short:   "create a new BackingServiceInstance",
		Long:    newBackingServiceInstanceLong,
		Example: fmt.Sprintf(newBackingServiceInstanceExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if options.complete(cmd, f); err != nil {
				kcmdutil.CheckErr(err)
			}

			if options.Client, _, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := options.Run(f); err != nil {
				fmt.Println("run err: ", err.Error())
			}
		},
	}

	cmd.Flags().StringVar(&options.DashboardUrl, "dashboard_url", "", "Dashboard URL")
	cmd.Flags().StringVar(&options.BackingServiceName, "backingservice_name", "", "BackingService GUID")
	cmd.Flags().StringVar(&options.BackingServicePlanGuid, "backingservice_plan_guid", "", "BackingService Plan GUID")
	
	return cmd
}

func (o *NewBackingServiceInstanceOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) == 0 {
		cmd.Help()
		return errors.New("must have at least 1 arguments")
	}

	o.Name = args[0]

	return nil
}

func (o *NewBackingServiceInstanceOptions) Run(f *clientcmd.Factory) error {
	backingServiceInstance := &backingserviceinstanceapi.BackingServiceInstance{}
	
	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	
	backingServiceInstance.Annotations = make(map[string]string)
	backingServiceInstance.Name = o.Name
	backingServiceInstance.GenerateName = o.Name
	
	backingServiceInstance.Spec.DashboardUrl = o.DashboardUrl
	backingServiceInstance.Spec.BackingServiceName = o.BackingServiceName
	backingServiceInstance.Spec.BackingServicePlanGuid = o.BackingServicePlanGuid
	backingServiceInstance.Spec.Parameters = make(map[string]string)
	
	//backingServiceInstance.Spec.Binding.BindUuid = 
	backingServiceInstance.Spec.InstanceBindDeploymentConfig = make(map[string]string)
	backingServiceInstance.Spec.Credential = make(map[string]string)
	
	//backingServiceInstance.Status = 
	
	_, err = o.Client.BackingServiceInstances(namespace).Create(backingServiceInstance)
	if err != nil {
		return err
	}

	return nil
}
