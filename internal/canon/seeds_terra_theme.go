package canon

// Add this to your existing canon seeding pass (called from RegisterAllRoutes or wherever you init canon).
func seedTerraTheme(seeds *[]DomainSeed) {
	*seeds = append(*seeds, DomainSeed{
		ID:     "domain.terra",
		Parent: "domain.null", // or whatever Terra’s parent is in your tree
		CSS: `
:root {
  --bg: #0b1220;
  --fg: #e6eef7;
  --muted: #17233a;
  --accent: #FFD60A; /* thin top border color */
  --link: #8bd3ff;
}

html, body {
  background-color: var(--bg) !important;
  color: var(--fg) !important;
}

a { color: var(--link); }
code, pre { background: var(--muted); }

 /* thin DIS indicator line at the very top */
body::before {
  content: "";
  position: fixed;
  inset: 0;
  border-top: 1px solid var(--accent);
  pointer-events: none;
  box-sizing: border-box;
  z-index: 2147483646; /* under your debug overlay at 99999 but high enough to sit on top */
}
`,
	})
}
