package resources

import (
	"strings"
	"testing"

	imagesv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/images/v1"
)

func image(id, name, organizationID string) *imagesv1.Image {
	return &imagesv1.Image{
		Meta:           &imagesv1.EntityMeta{Id: id},
		Name:           name,
		OrganizationId: organizationID,
	}
}

func TestSelectImageByNameFindsTheOnlyMatch(t *testing.T) {
	images := []*imagesv1.Image{image("a", "devcontainer", "org-1"), image("b", "codex", "org-1")}
	found, diagnostic := selectImageByName(images, "codex", "org-1")
	if diagnostic != "" {
		t.Fatalf("unexpected diagnostic: %s", diagnostic)
	}
	if found.GetMeta().GetId() != "b" {
		t.Fatalf("got image %s, want b", found.GetMeta().GetId())
	}
}

// The list carries the organization's own images and every public one, so a
// name can legitimately appear twice. The organization's own must win.
func TestSelectImageByNameOwnImageShadowsPublic(t *testing.T) {
	images := []*imagesv1.Image{image("public", "codex", "platform-org"), image("own", "codex", "org-1")}
	found, diagnostic := selectImageByName(images, "codex", "org-1")
	if diagnostic != "" {
		t.Fatalf("unexpected diagnostic: %s", diagnostic)
	}
	if found.GetMeta().GetId() != "own" {
		t.Fatalf("got image %s, want own", found.GetMeta().GetId())
	}
}

func TestSelectImageByNameReportsAmbiguity(t *testing.T) {
	images := []*imagesv1.Image{image("one", "codex", "platform-org"), image("two", "codex", "other-org")}
	found, diagnostic := selectImageByName(images, "codex", "org-1")
	if found != nil {
		t.Fatalf("got image %s, want none", found.GetMeta().GetId())
	}
	for _, want := range []string{"ambiguous", "one", "two"} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("diagnostic %q does not mention %q", diagnostic, want)
		}
	}
}

func TestSelectImageByNameReportsMiss(t *testing.T) {
	found, diagnostic := selectImageByName([]*imagesv1.Image{image("a", "devcontainer", "org-1")}, "codex", "org-1")
	if found != nil {
		t.Fatal("got an image, want none")
	}
	if !strings.Contains(diagnostic, "codex") {
		t.Fatalf("diagnostic %q does not name the image looked for", diagnostic)
	}
}
