import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center rounded-sm px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-1 focus:ring-[var(--color-primary-cyan)]",
  {
    variants: {
      variant: {
        default:
          "border border-[var(--color-primary-cyan)]/30 bg-[var(--color-primary-cyan)]/10 text-[var(--color-primary-cyan)] shadow hover:bg-[var(--color-primary-cyan)]/20",
        secondary:
          "border border-[var(--color-border-dim)] bg-[var(--color-surface-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-elevated)]/80",
        destructive:
          "border border-[var(--color-danger-rose)]/30 bg-[var(--color-danger-rose)]/10 text-[var(--color-danger-rose)] shadow hover:bg-[var(--color-danger-rose)]/20",
        outline:
          "border border-[var(--color-border-dim)] text-[var(--color-text-secondary)] hover:border-[var(--color-border-active)]",
        success:
          "border border-[var(--color-success-teal)]/30 bg-[var(--color-success-teal)]/10 text-[var(--color-success-teal)] shadow hover:bg-[var(--color-success-teal)]/20",
        warning:
          "border border-[var(--color-manufacturing-amber)]/30 bg-[var(--color-manufacturing-amber)]/10 text-[var(--color-manufacturing-amber)] shadow hover:bg-[var(--color-manufacturing-amber)]/20",
        info:
          "border border-[var(--color-science-blue)]/30 bg-[var(--color-science-blue)]/10 text-[var(--color-science-blue)] shadow hover:bg-[var(--color-science-blue)]/20",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  )
}

export { Badge, badgeVariants }
