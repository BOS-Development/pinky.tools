"use client"

import { Toaster as Sonner, toast } from "sonner"

type ToasterProps = React.ComponentProps<typeof Sonner>

const Toaster = ({ ...props }: ToasterProps) => {
  return (
    <Sonner
      theme="dark"
      className="toaster group"
      toastOptions={{
        classNames: {
          toast:
            "group toast group-[.toaster]:bg-[var(--color-bg-panel)] group-[.toaster]:text-[var(--color-text-primary)] group-[.toaster]:border-[var(--color-border-active)] group-[.toaster]:shadow-[var(--glow-cyan-sm)]",
          description: "group-[.toast]:text-[var(--color-text-secondary)]",
          actionButton:
            "group-[.toast]:bg-[var(--color-primary-cyan)] group-[.toast]:text-[var(--color-bg-void)]",
          cancelButton:
            "group-[.toast]:bg-[var(--color-surface-elevated)] group-[.toast]:text-[var(--color-text-secondary)]",
          success: "group-[.toaster]:border-[var(--color-success-teal)]/30",
          error: "group-[.toaster]:border-[var(--color-danger-rose)]/30",
        },
      }}
      {...props}
    />
  )
}

export { Toaster, toast }
