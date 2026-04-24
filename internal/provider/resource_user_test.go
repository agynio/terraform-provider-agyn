package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynUser_basic(t *testing.T) {
	oidcSubject := fmt.Sprintf("%s@example.com", acctest.RandomWithPrefix("tf-acc-user"))
	name := "Terraform acceptance user"
	photoURL := "https://example.com/user.png"
	username := acctest.RandomWithPrefix("tf-acc-username")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccUserPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynUserConfig(oidcSubject, name, photoURL, username, "none"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_user.test", "oidc_subject", oidcSubject),
					resource.TestCheckResourceAttr("agyn_user.test", "name", name),
					resource.TestCheckResourceAttr("agyn_user.test", "photo_url", photoURL),
					resource.TestCheckResourceAttr("agyn_user.test", "username", username),
					resource.TestCheckResourceAttr("agyn_user.test", "cluster_role", "none"),
					resource.TestCheckResourceAttrSet("agyn_user.test", "identity_id"),
				),
			},
		},
	})
}

func TestAccAgynUser_update(t *testing.T) {
	oidcSubject := fmt.Sprintf("%s@example.com", acctest.RandomWithPrefix("tf-acc-user"))
	name := "Terraform acceptance user"
	photoURL := "https://example.com/user.png"
	username := acctest.RandomWithPrefix("tf-acc-username")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccUserPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynUserConfig(oidcSubject, name, photoURL, username, "none"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_user.test", "cluster_role", "none"),
					resource.TestCheckResourceAttrSet("agyn_user.test", "identity_id"),
				),
			},
			{
				Config: testAccAgynUserConfig(oidcSubject, name, photoURL, username, "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_user.test", "cluster_role", "admin"),
					resource.TestCheckResourceAttrSet("agyn_user.test", "identity_id"),
				),
			},
		},
	})
}

func TestAccAgynUser_import(t *testing.T) {
	oidcSubject := fmt.Sprintf("%s@example.com", acctest.RandomWithPrefix("tf-acc-user"))
	name := "Terraform acceptance user"
	photoURL := "https://example.com/user.png"
	username := acctest.RandomWithPrefix("tf-acc-username")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccUserPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynUserConfig(oidcSubject, name, photoURL, username, "admin"),
			},
			{
				ResourceName:      "agyn_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv("AGYN_API_TOKEN") == "" {
		t.Skip("AGYN_API_TOKEN must be set for user acceptance tests")
	}
}

func testAccAgynUserConfig(oidcSubject, name, photoURL, username, clusterRole string) string {
	nameLine := ""
	if name != "" {
		nameLine = fmt.Sprintf("\n\t  name         = %q", name)
	}
	photoLine := ""
	if photoURL != "" {
		photoLine = fmt.Sprintf("\n\t  photo_url    = %q", photoURL)
	}
	usernameLine := ""
	if username != "" {
		usernameLine = fmt.Sprintf("\n\t  username     = %q", username)
	}
	clusterLine := ""
	if clusterRole != "" {
		clusterLine = fmt.Sprintf("\n\t  cluster_role = %q", clusterRole)
	}

	return fmt.Sprintf(`
%s

resource "agyn_user" "test" {
	  oidc_subject = %q%s%s%s%s
}
`, testAccProviderConfig(), oidcSubject, nameLine, photoLine, usernameLine, clusterLine)
}
