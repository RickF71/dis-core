package canon

func seedUSADomainTheme(seeds *[]DomainSeed) {
	*seeds = append(*seeds, DomainSeed{
		ID:     "domain.usa",
		Parent: "domain.terra",
		CSS: `
:root {
  --accent: #B22234;    /* deeper red */
  --link:   #3C6EE1;    /* calmer blue link */
  --badge:  #3C3B6E;    /* blue for small UI badges if you use them */
}

/* keep Terra base; just override what we need */
a { color: var(--link); }

/* make the top indicator line USA-red */
body::before {
  border-top: 1px solid var(--accent);
}
`,
	})
}
