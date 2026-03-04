import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./packages/**/*.{js,ts,jsx,tsx,mdx}",
    "./lib/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: {
          void: "var(--color-bg-void)",
          panel: "var(--color-bg-panel)",
          elevated: "var(--color-surface-elevated)",
        },
        border: {
          dim: "var(--color-border-dim)",
          active: "var(--color-border-active)",
        },
        primary: {
          DEFAULT: "var(--color-primary-cyan)",
          muted: "var(--color-cyan-muted)",
        },
        amber: {
          manufacturing: "var(--color-manufacturing-amber)",
        },
        blue: {
          science: "var(--color-science-blue)",
        },
        teal: {
          success: "var(--color-success-teal)",
        },
        rose: {
          danger: "var(--color-danger-rose)",
        },
        status: {
          "success-tint": "var(--color-success-tint)",
          "warning-tint": "var(--color-warning-tint)",
          "error-tint": "var(--color-error-tint)",
          "info-tint": "var(--color-info-tint)",
          "neutral-tint": "var(--color-neutral-tint)",
        },
        category: {
          violet: "var(--color-category-violet)",
          pink: "var(--color-category-pink)",
          orange: "var(--color-category-orange)",
          teal: "var(--color-category-teal)",
          slate: "var(--color-category-slate)",
        },
        "accent-blue": {
          DEFAULT: "var(--color-accent-blue)",
          hover: "var(--color-accent-blue-hover)",
          muted: "var(--color-accent-blue-muted)",
        },
        "bg-manufacturing": "var(--color-bg-manufacturing)",
        "bg-science": "var(--color-bg-science)",
        "bg-warning": "var(--color-bg-warning)",
        text: {
          primary: "var(--color-text-primary)",
          secondary: "var(--color-text-secondary)",
          muted: "var(--color-text-muted)",
          emphasis: "var(--color-text-emphasis)",
          heading: "var(--color-heading)",
          "data-value": "var(--color-data-value)",
        },
        // shadcn/ui semantic tokens
        foreground: "hsl(var(--foreground))",
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        interactive: {
          hover: "var(--color-interactive-hover)",
          active: "var(--color-interactive-active)",
          selected: "var(--color-interactive-selected)",
          disabled: "var(--color-interactive-disabled)",
        },
        overlay: {
          subtle: "var(--color-overlay-subtle)",
          medium: "var(--color-overlay-medium)",
          strong: "var(--color-overlay-strong)",
        },
      },
      borderRadius: {
        none: "0",
        DEFAULT: "var(--radius-default, 0.125rem)",
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
        full: "9999px",
      },
      boxShadow: {
        "glow-sm": "var(--glow-cyan-sm)",
        "glow-md": "var(--glow-cyan-md)",
        "glow-lg": "var(--glow-cyan-lg)",
        "glow-text": "var(--glow-text-heading)",
      },
      fontFamily: {
        sans: ["var(--font-geist-sans)", "system-ui", "sans-serif"],
        mono: ["var(--font-jetbrains-mono)", "var(--font-geist-mono)", "monospace"],
        display: ["var(--font-exo2)", "var(--font-geist-sans)", "system-ui", "sans-serif"],
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
      },
    },
  },
  plugins: [],
};

export default config;
