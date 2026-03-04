import { type LucideIcon } from 'lucide-react';

type IconProps = {
  icon: LucideIcon;
  size?: number;
  color?: string;
  className?: string;
};

/**
 * Wrapper component for Lucide React icons.
 * Provides consistent sizing and optional className support.
 *
 * Usage:
 *   import { Rocket } from 'lucide-react';
 *   <Icon icon={Rocket} size={20} color="var(--color-primary-cyan)" />
 */
export default function Icon({ icon: LucideIcon, size = 20, color, className }: IconProps) {
  return (
    <span className={`inline-flex items-center justify-center leading-none ${className ?? ''}`} style={{ color: color ?? 'inherit' }}>
      <LucideIcon size={size} />
    </span>
  );
}
