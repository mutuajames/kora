package api

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/kora/auth"
	"github.com/yourorg/kora/doctype"
)

// --- Auth Providers ---

// HandleAuthProviders returns enabled authentication providers.
// Public endpoint — no auth required.
func (h *Handler) HandleAuthProviders(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Data: map[string]any{
			"providers": []map[string]any{
				{"name": "password", "label": "Email & Password"},
			},
		},
	})
}

// --- System Doctype ---

// ReferenceInfo describes a doctype that links to the current doctype via a Link field.
type ReferenceInfo struct {
	Doctype   string `json:"doctype"`
	Fieldname string `json:"fieldname"`
	Label     string `json:"label"`
}

// SystemDoctypeResponse is the full schema response for a single DocType.
type SystemDoctypeResponse struct {
	DocType      *doctype.DocType              `json:"doctype"`
	Workflow     *WorkflowResponse              `json:"workflow,omitempty"`
	Permissions  map[string]bool               `json:"permissions"`
	Transitions  []doctype.WorkflowTransition  `json:"transitions,omitempty"`
	ReferencedBy []ReferenceInfo               `json:"referenced_by,omitempty"`
}

// WorkflowResponse holds the workflow definition for a DocType.
type WorkflowResponse struct {
	States      []doctype.WorkflowState      `json:"states"`
	Transitions []doctype.WorkflowTransition `json:"transitions"`
	StateField  string                       `json:"state_field"`
}

// HandleSystemDoctype returns the full DocType schema with workflow and permissions.
// GET /api/system/doctype/:doctype
// Optional query param: ?state=current_state to get available transitions.
func (h *Handler) HandleSystemDoctype(c *gin.Context) {
	doctypeName := c.Param("doctype")
	reg := h.siteRegistry(c)
	dt := reg.Get(doctypeName)
	if dt == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: map[string]string{"message": "DocType not found: " + doctypeName},
		})
		return
	}

	resp := SystemDoctypeResponse{
		DocType:     dt,
		Permissions: getUserPermissions(reg, c, doctypeName),
	}

	// Compute which doctypes link to this one (back-references).
	resp.ReferencedBy = findReferencingDoctypes(reg, doctypeName)

	// Attach workflow data if this doctype has one.
	if reg.Workflows.Has(doctypeName) {
		wf := reg.Workflows.Get(doctypeName)
		resp.Workflow = &WorkflowResponse{
			States:      wf.States,
			Transitions: wf.Transitions,
			StateField:  wf.WorkflowStateField,
		}

		// Compute available transitions for the current state and user.
		currentState := c.Query("state")
		if currentState != "" {
			userRole := c.GetString("user_role")
			if userRole == "" {
				userRole = doctype.AdminRole
			}
			doc := doctype.NewDocument(doctypeName)
			for key, vals := range c.Request.URL.Query() {
				if key != "state" && len(vals) > 0 {
					doc.Set(key, vals[0])
				}
			}
			resp.Transitions = reg.Workflows.GetAvailableTransitions(doctypeName, currentState, userRole, doc)
		}
	}

	c.JSON(http.StatusOK, Response{Data: resp})
}

// getUserPermissions returns a map of operation → allowed for the current user on a doctype.
func getUserPermissions(reg *doctype.Registry, c *gin.Context, dt string) map[string]bool {
	userRoles := c.GetStringSlice("user_roles")
	// If no roles set, return full access (bootstrapping / system user).
	if len(userRoles) == 0 {
		return map[string]bool{
			"read": true, "write": true, "create": true, "delete": true,
			"submit": true, "cancel": true, "amend": true,
			"export": true, "import": true, "report": true,
		}
	}
	ops := []string{"read", "write", "create", "delete", "submit", "cancel", "amend", "export", "import", "report"}
	perms := make(map[string]bool, len(ops))
	for _, op := range ops {
		allowed, _ := reg.CanUser(userRoles, dt, op)
		perms[op] = allowed
	}
	return perms
}

// --- System Navigation ---

// NavigationResponse is the full navigation config for the SPA sidebar.
type NavigationResponse struct {
	Modules  []ModuleGroup `json:"modules"`
	Branding Branding      `json:"branding"`
	User     UserInfo      `json:"user"`
}

// ModuleGroup is a group of DocTypes under a module.
type ModuleGroup struct {
	Module   string           `json:"module"`
	Label    string           `json:"label"`
	DocTypes []DocTypeNavItem `json:"doctypes"`
}

// DocTypeNavItem is a single DocType entry in the navigation.
type DocTypeNavItem struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Icon    string `json:"icon,omitempty"`
	IsChild bool   `json:"is_child"`
}

// AppBranding is the global branding config (set from common config at startup).
var AppBranding = Branding{AppName: "Kora", PrimaryColor: "#2563eb"}

// Branding holds per-site branding configuration.
type Branding struct {
	AppName      string `json:"app_name"`
	PrimaryColor string `json:"primary_color"`
}

// UserInfo is the current user's public info for the UI.
type UserInfo struct {
	Name     string   `json:"name"`
	FullName string   `json:"full_name"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// HandleSystemNavigation returns the navigation config (sidebar, branding, user).
// GET /api/system/navigation
func (h *Handler) HandleSystemNavigation(c *gin.Context) {
	reg := h.siteRegistry(c)
	doctypes := reg.All()

	// Group by module, skip child tables.
	moduleMap := make(map[string][]DocTypeNavItem)
	for _, dt := range doctypes {
		if dt.IsChildTable {
			continue
		}
		module := dt.Module
		if module == "" {
			module = "System"
		}
		moduleMap[module] = append(moduleMap[module], DocTypeNavItem{
			Name:    dt.Name,
			Label:   dt.Name,
			IsChild: false,
		})
	}

	// Sort modules deterministically.
	moduleNames := make([]string, 0, len(moduleMap))
	for m := range moduleMap {
		moduleNames = append(moduleNames, m)
	}
	sort.Strings(moduleNames)

	var modules []ModuleGroup
	for _, m := range moduleNames {
		items := moduleMap[m]
		// Sort doctypes within module.
		sort.Slice(items, func(i, j int) bool {
			return items[i].Label < items[j].Label
		})
		modules = append(modules, ModuleGroup{
			Module:   m,
			Label:    m,
			DocTypes: items,
		})
	}

	// Extract user info from context (set by SiteGuard/AuthMiddleware).
	user := UserInfo{}
	if userObj, exists := c.Get("user_obj"); exists {
		if u, ok := userObj.(*auth.User); ok {
			user.Name = u.Name
			user.FullName = u.FullName
			user.Email = u.Email
			user.Roles = u.Roles
		}
	}
	// Fallback: read individual context values.
	if user.Name == "" {
		user.Name = c.GetString("user")
	}
	if len(user.Roles) == 0 {
		user.Roles = c.GetStringSlice("user_roles")
	}
	if user.Email == "" {
		user.Email = c.GetString("user_email")
		if user.Email == "" {
			user.Email = c.GetString("user")
		}
	}
	if user.FullName == "" {
		user.FullName = c.GetString("user_full_name")
		if user.FullName == "" {
			user.FullName = user.Name
		}
	}

	branding := AppBranding

	c.JSON(http.StatusOK, Response{
		Data: NavigationResponse{
			Modules:  modules,
			Branding: branding,
			User:     user,
		},
	})
}

// findReferencingDoctypes returns a list of doctypes that have Link fields pointing to targetDoctype.
func findReferencingDoctypes(reg *doctype.Registry, targetDoctype string) []ReferenceInfo {
	var refs []ReferenceInfo
	for _, dt := range reg.All() {
		if dt.IsChildTable {
			continue
		}
		for _, f := range dt.Fields {
			if (f.Fieldtype == "Link" || f.Fieldtype == "Dynamic Link") && f.Options == targetDoctype {
				refs = append(refs, ReferenceInfo{
					Doctype:   dt.Name,
					Fieldname: f.Fieldname,
					Label:     f.Label,
				})
			}
		}
	}
	return refs
}

// RegisterSystemRoutes registers system endpoints on the given API group.
func RegisterSystemRoutes(apiGroup *gin.RouterGroup, handler *Handler) {
	system := apiGroup.Group("/system")
	{
		system.GET("/doctype/:doctype", handler.HandleSystemDoctype)
		system.GET("/navigation", handler.HandleSystemNavigation)
	}
}
