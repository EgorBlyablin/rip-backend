// Add module declarations so imports like "../assets/search.svg?react" are recognized by TypeScript.

declare module '*.svg' {
	// Default export is the URL string (for raw imports)
	const src: string;
	export default src;

	// Named export for React component (common pattern)
	export const ReactComponent: React.FunctionComponent<
		React.SVGProps<SVGSVGElement> & { title?: string }
	>;
}

declare module '*.svg?react' {
	// When importing with ?react we expect a default React component:
	import * as React from 'react';
	const ReactComponent: React.FunctionComponent<
		React.SVGProps<SVGSVGElement> & { title?: string }
	>;
	export default ReactComponent;
}
