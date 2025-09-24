package admin_test

import (
	"strings"
	"testing"

	"github.com/zy84338719/filecodebox/internal/models"
	admin "github.com/zy84338719/filecodebox/internal/services/admin"
)

func TestUpdateUserWithParamsHashesPasswordAndUpdatesRole(t *testing.T) {
	svc, repo, _ := setupAdminTestService(t)

	user := &models.User{
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "legacy",
		Role:         "user",
		Status:       "active",
	}
	if err := repo.User.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	newEmail := "alice-updated@example.com"
	newPassword := "StrongPass123"
	isAdmin := true
	isActive := false

	params := admin.UserUpdateParams{
		Email:    &newEmail,
		Password: &newPassword,
		IsAdmin:  &isAdmin,
		IsActive: &isActive,
	}

	if err := svc.UpdateUserWithParams(user.ID, params); err != nil {
		t.Fatalf("UpdateUserWithParams returned error: %v", err)
	}

	updated, err := repo.User.GetByID(user.ID)
	if err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}

	if updated.Email != newEmail {
		t.Fatalf("expected email to be updated, got %s", updated.Email)
	}
	if updated.Role != "admin" {
		t.Fatalf("expected role to be admin, got %s", updated.Role)
	}
	if updated.Status != "inactive" {
		t.Fatalf("expected status to be inactive, got %s", updated.Status)
	}
	if updated.PasswordHash == newPassword || updated.PasswordHash == "" {
		t.Fatalf("expected password to be hashed, got %s", updated.PasswordHash)
	}
}

func TestUpdateUserWithParamsDuplicateEmail(t *testing.T) {
	svc, repo, _ := setupAdminTestService(t)

	user1 := &models.User{Username: "bob", Email: "bob@example.com", PasswordHash: "hash"}
	if err := repo.User.Create(user1); err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}

	user2 := &models.User{Username: "carol", Email: "carol@example.com", PasswordHash: "hash"}
	if err := repo.User.Create(user2); err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	duplicateEmail := "carol@example.com"
	params := admin.UserUpdateParams{Email: &duplicateEmail}

	err := svc.UpdateUserWithParams(user1.ID, params)
	if err == nil {
		t.Fatal("expected duplicate email error, got nil")
	}
	if !strings.Contains(err.Error(), "该邮箱已被使用") {
		t.Fatalf("unexpected error message: %v", err)
	}

	// ensure original email unchanged
	reloaded, err := repo.User.GetByID(user1.ID)
	if err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if reloaded.Email != "bob@example.com" {
		t.Fatalf("expected email to remain unchanged, got %s", reloaded.Email)
	}
}

func TestUpdateUserWithParamsMissingUser(t *testing.T) {
	svc, _, _ := setupAdminTestService(t)

	err := svc.UpdateUserWithParams(9999, admin.UserUpdateParams{})
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
	if !strings.Contains(err.Error(), "用户不存在") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
