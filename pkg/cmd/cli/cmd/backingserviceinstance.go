package cmd

import (
	"errors"
	"fmt"
	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
	"io"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	newBackingServiceInstanceLong = `
Create a new BackingServiceInstance

`
	newBackingServiceInstanceExample = `# Create a new backingserviceinstance with [name BackingServiceGuidrname BackingServicePlanGuid]
  $ %[1]s  mysql_BackingServiceInst  --backingserviceguid="bs-01023" --planid="ab98df31"`
)

type NewBackingServiceInstanceOptions struct {
	DashboardUrl           string
	BackingServiceGuid     string
	BackingServicePlanGuid string

	Client client.Interface

	Out io.Writer
}

func NewCmdBackingServiceInstance(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &NewBackingServiceInstanceOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     "new-BackingServiceInst NAME --backingservice_guid=BackingServiceGuid --plan_guid=BackingServicePlanGuid --dashboard_url=DashboardUrl",
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

			if err := options.Run(); err != nil {
				fmt.Println("run err %s", err.Error())
			}
		},
	}

	cmd.Flags().StringVar(&options.DashboardUrl, "dashboard_url", "", "Dashboard URL")
	cmd.Flags().StringVar(&options.BackingServiceGuid, "backingservice_guid", "", "BackingService GUID")
	cmd.Flags().StringVar(&options.BackingServicePlanGuid, "plan_guid", "", "BackingService Plan GUID")

	return cmd
}

func (o *NewBackingServiceInstanceOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) < 3 {
		cmd.Help()
		return errors.New("must have at least 3 arguments")
	}

	//o.Name = args[0]

	return nil
}

func (o *NewBackingServiceInstanceOptions) Run() error {
	backingServiceInstance := &backingserviceinstanceapi.BackingServiceInstance{}
	//backingServiceInstance.Spec.Provisioning.DashboardUrl = o.DashboardUrl
	//backingServiceInstance.Spec.Provisioning.BackingServiceGuid = o.BackingServiceGuid
	//backingServiceInstance.Spec.Provisioning.BackingServicePlanGuid = o.BackingServicePlanGuid
	//backingServiceInstance.Spec.Provisioning.Parameters = make(map[string]string)
	//
	////backingServiceInstance.Spec.Binding.BindUuid =
	//backingServiceInstance.Spec.Binding.InstanceBindDeploymentConfig = make(map[string]string)
	//backingServiceInstance.Spec.Binding.Credential = make(map[string]string)

	//backingServiceInstance.Status =

	_, err := o.Client.BackingServiceInstances().Create(backingServiceInstance)
	if err != nil {
		return err
	}

	return nil
}
