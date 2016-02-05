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
Create a new backingserviceinstance

`
	newBackingServiceInstanceExample = `# Create a new backingserviceinstance with [name BackingServiceGuidrname BackingServicePlanGuid]
  $ %[1]s  mysql_backingserviceinstance  --backingserviceguid="bs-01023"  --planid="ab98df31"`
)

type NewBackingServiceInstanceOptions struct {
	BackingServiceGuid     string
	BackingServicePlanGuid string

	Client client.Interface

	Out io.Writer
}

func NewCmdBackingServiceInstance(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &NewBackingServiceInstanceOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     "new-backingserviceinstance NAME --backingserviceguid=BackingServiceGuid --planid=BackingServicePlanGuid",
		Short:   "create a new backingserviceinstance",
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

	cmd.Flags().StringVar(&options.BackingServiceGuid, "backingserviceguid", "", "BackingService GUID")
	cmd.Flags().StringVar(&options.BackingServicePlanGuid, "planid", "", "BackingService PlanId")

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
	backingServiceInstance.Spec.BackingServiceGuid = o.BackingServiceGuid
	backingServiceInstance.Spec.BackingServicePlanGuid = o.BackingServicePlanGuid
	backingServiceInstance.Annotations = make(map[string]string)

	_, err := o.Client.BackingServiceInstances().Create(backingServiceInstance)
	if err != nil {
		return err
	}

	return nil
}
