package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/database"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services"
)

var rootCmd = &cobra.Command{
	Use:   "filecodebox",
	Short: "FileCodeBox CLI tools",
	Long:  "Command-line tools for FileCodeBox (admin management, maintenance, etc)",
}

var cfgPath string
var dataPath string

func init() {
	// global flags
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "Path to config.yaml to load")
	rootCmd.PersistentFlags().StringVar(&dataPath, "data-path", "", "Override data path (overrides DATA_PATH env)")
}

func printVersion() {
	buildInfo := models.GetBuildInfo()
	fmt.Printf("FileCodeBox %s\nCommit: %s\nBuilt: %s\nGo Version: %s\n", buildInfo.Version, buildInfo.GitCommit, buildInfo.BuildTime, runtime.Version())
}

// Execute executes the root cobra command
func Execute() {
	if len(os.Args) == 2 && (os.Args[1] == "-version" || os.Args[1] == "--version") {
		printVersion()
		return
	}

	// add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(adminCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// helper to initialize services that need DB
func initServices() (*repository.RepositoryManager, *services.AdminService, error) {
	manager := config.InitManager()

	// If a config file is provided, prefer it (and let ConfigManager track managed keys)
	if cfgPath != "" {
		if err := manager.LoadFromYAML(cfgPath); err != nil {
			// try to continue, but log to stderr
			fmt.Fprintln(os.Stderr, "warning: failed to load config file:", err)
		}
	}

	// Override data path if flag provided
	if dataPath != "" {
		manager.Base.DataPath = dataPath
	}

	// init DB
	db, err := database.InitWithManager(manager)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init database: %w", err)
	}

	repoMgr := repository.NewRepositoryManager(db)
	storageService := services.NewAdminService(repoMgr, manager, nil) // placeholder: admin.NewService expects storageService; we pass nil where not needed
	// Actually services.NewAdminService signature in services package returns admin.Service alias; reuse admin.Service
	adminService := storageService

	return repoMgr, adminService, nil
}

// placeholders to be implemented in separate files
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin user management commands",
}
