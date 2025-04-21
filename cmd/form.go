package cmd

import (
	"github.com/spf13/cobra"
)

func NewFormCommand() *cobra.Command {

	FormCmd := &cobra.Command{
		Use:   "Form",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly Form a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// form := ui.RenderForm()

			// if err := form.Run(); err != nil {
			// 	fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
			// }

			// fmt.Print(form.Get("burger"))
			// fmt.Println(ui.Burger)

			return nil
		},
	}

	return FormCmd
}
