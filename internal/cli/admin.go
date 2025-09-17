package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	_ "github.com/zy84338719/filecodebox/internal/config"
	_ "github.com/zy84338719/filecodebox/internal/database"
	"github.com/zy84338719/filecodebox/internal/models"
	_ "github.com/zy84338719/filecodebox/internal/repository"
	_ "github.com/zy84338719/filecodebox/internal/services"
	"golang.org/x/term"
)

var adminCreateCmd = &cobra.Command{
	Use:   "create [username] [password] [email]",
	Short: "Create admin user",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoMgr, adminSvc, err := initServices()
		if err != nil {
			return err
		}
		// repoMgr has no Close method; repository manager lifecycle handled by program

		username := args[0]
		password := args[1]
		email := args[2]
		// 如果当前数据库没有用户，提示友好引导；用户可以在网页上完成初始化
		if repoMgr != nil {
			if cnt, err := repoMgr.User.Count(); err == nil && cnt == 0 {
				// 检测到没有用户
				force, _ := cmd.Flags().GetBool("force")
				if !force {
					fmt.Println("未检测到任何用户。注意：首次通过网页初始化时，第一位创建的用户将自动成为管理员。")
					fmt.Println("你可以通过浏览器访问管理后台完成自助初始化 (例如: http://localhost:12345/setup 或 http://localhost:12345/admin/setup)，")
					fmt.Println("或者如果你确实想要通过 CLI 创建第一个用户，请重新运行本命令并加上 --force 标志以强制创建。")
					return nil
				}
			}
		}

		_, err = adminSvc.CreateUser(username, email, password, "admin", "admin", "active")
		if err != nil {
			return err
		}
		fmt.Println("admin user created")
		return nil
	},
}

func init() {
	// allow forcing first-user creation when DB empty
	adminCreateCmd.Flags().BoolP("force", "f", false, "Force create first user even if DB has no users (first user will be admin)")
}

var adminResetCmd = &cobra.Command{
	Use:   "reset [userID|username] [newPassword]",
	Short: "Reset user password by ID or username",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoMgr, adminSvc, err := initServices()
		if err != nil {
			return err
		}
		// repoMgr has no Close method; repository manager lifecycle handled by program

		// 如果没有用户，友好提示并引导到网页初始化
		if repoMgr != nil {
			if cnt, err := repoMgr.User.Count(); err == nil && cnt == 0 {
				fmt.Println("未检测到任何用户。首次用户可通过网页自助初始化，第一位用户将成为管理员。请先在网页上完成初始化或使用 CLI 创建第一个用户（见 create --force）")
				return nil
			}
		}

		identifier := args[0]
		var userID uint
		if id64, err := strconv.ParseUint(identifier, 10, 64); err == nil {
			userID = uint(id64)
		} else {
			// treat as username
			u, err := repoMgr.User.GetByUsername(identifier)
			if err != nil {
				return fmt.Errorf("找不到用户: %w", err)
			}
			userID = u.ID
		}

		var newPass string
		if len(args) == 2 {
			newPass = args[1]
		} else {
			// prompt for password (hidden)
			fmt.Print("New password: ")
			pwBytes, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				return err
			}
			newPass = strings.TrimSpace(string(pwBytes))
			if newPass == "" {
				// if empty, optionally prompt once more via stdin
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Password cannot be empty, please enter again: ")
				line, _ := reader.ReadString('\n')
				newPass = strings.TrimSpace(line)
				if newPass == "" {
					return fmt.Errorf("password cannot be empty")
				}
			}
		}

		// Build models.User with ID and new password (service will handle update semantics)
		user := models.User{}
		// set ID via embedded gorm.Model field
		// using reflection-free assignment
		// gorm.Model exposes ID field; set using map-style assignment
		// but we set via simple field since it's promoted
		user.ID = userID
		// Store new password in PasswordHash field temporarily; repository will map it to password_hash
		user.PasswordHash = newPass
		return adminSvc.UpdateUser(user)
	},
}

var adminListCmd = &cobra.Command{
	Use:   "list [page] [pageSize]",
	Short: "List users",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoMgr, adminSvc, err := initServices()
		if err != nil {
			return err
		}
		// repoMgr has no Close method; repository manager lifecycle handled by program

		page := 1
		pageSize := 20
		if len(args) >= 1 {
			if p, err := strconv.Atoi(args[0]); err == nil {
				page = p
			}
		}
		if len(args) == 2 {
			if ps, err := strconv.Atoi(args[1]); err == nil {
				pageSize = ps
			}
		}

		// 如果没有用户，友好提示并引导到网页初始化
		if repoMgr != nil {
			if cnt, err := repoMgr.User.Count(); err == nil && cnt == 0 {
				fmt.Println("未检测到任何用户。首次用户可通过网页自助初始化，第一位用户将成为管理员。请先在网页上完成初始化（例如访问 /setup 或 /admin/setup）或使用 create --force 在 CLI 创建第一个用户。")
				return nil
			}
		}

		users, total, err := adminSvc.GetUsers(page, pageSize, "")
		if err != nil {
			return err
		}
		fmt.Printf("total: %d\n", total)
		for _, u := range users {
			fmt.Printf("%d: %s (%s)\n", u.ID, u.Username, u.Email)
		}
		return nil
	},
}

func init() {
	adminCmd.AddCommand(adminCreateCmd)
	adminCmd.AddCommand(adminResetCmd)
	adminCmd.AddCommand(adminListCmd)
}
