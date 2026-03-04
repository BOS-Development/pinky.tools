import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-sm text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--color-primary-cyan)] disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default:
          "bg-[var(--color-primary-cyan)] text-[var(--color-bg-void)] shadow hover:bg-[var(--color-cyan-muted)] font-semibold",
        destructive:
          "bg-[var(--color-danger-rose)] text-[var(--color-text-primary)] shadow-sm hover:bg-[var(--color-danger-rose)]/90",
        outline:
          "border border-[var(--color-border-active)] bg-transparent text-[var(--color-text-primary)] shadow-sm hover:bg-[var(--color-surface-elevated)] hover:border-[var(--color-primary-cyan)]",
        secondary:
          "bg-[var(--color-surface-elevated)] text-[var(--color-text-primary)] shadow-sm hover:bg-[var(--color-surface-elevated)]/80",
        ghost:
          "text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-elevated)] hover:text-[var(--color-text-primary)]",
        link: "text-[var(--color-primary-cyan)] underline-offset-4 hover:underline",
        neocom:
          "border border-[var(--color-primary-cyan)] bg-transparent text-[var(--color-primary-cyan)] shadow-glow-sm hover:bg-[var(--color-primary-cyan)]/10 hover:shadow-glow-md font-semibold tracking-wide",
      },
      size: {
        default: "h-9 px-4 py-2",
        sm: "h-8 rounded-sm px-3 text-xs",
        lg: "h-10 rounded-sm px-8",
        icon: "h-9 w-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button, buttonVariants }
