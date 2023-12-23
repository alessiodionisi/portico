package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alessiodionisi/portico/controller"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	gclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
	ginformers "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions"
)

func newController() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Run the Portico controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runController(); err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}

			return nil
		},
	}

	return cmd
}

func runController() error {
	ctx := context.Background()

	config, err := clientcmd.BuildConfigFromFlags("", "/Users/alessiodionisi/.kube/config")
	if err != nil {
		return fmt.Errorf("error building kubeconfig: %w", err)
	}

	gatewayClientset, err := gclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error building gateway clientset: %w", err)
	}

	gatewayInformerFactory := ginformers.NewSharedInformerFactory(gatewayClientset, time.Second*30)

	ctrl := controller.New(
		ctx,
		gatewayClientset,
		gatewayInformerFactory.Gateway().V1().GatewayClasses(),
		gatewayInformerFactory.Gateway().V1().Gateways(),
	)

	gatewayInformerFactory.Start(ctx.Done())

	if err := ctrl.Run(ctx, 2); err != nil {
		return fmt.Errorf("error running controller: %w", err)
	}

	return nil
}
