import { type LucideIcon } from 'lucide-react';
import Box from '@mui/material/Box';
import type { SxProps, Theme } from '@mui/material/styles';

type IconProps = {
  icon: LucideIcon;
  size?: number;
  color?: string;
  sx?: SxProps<Theme>;
};

/**
 * Wrapper component for Lucide React icons.
 * Provides consistent sizing and MUI sx prop support for gradual migration
 * from @mui/icons-material to Lucide.
 *
 * Usage:
 *   import { Rocket } from 'lucide-react';
 *   <Icon icon={Rocket} size={20} color="#00d4ff" />
 */
export default function Icon({ icon: LucideIcon, size = 20, color, sx }: IconProps) {
  return (
    <Box
      component="span"
      sx={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: color ?? 'inherit',
        lineHeight: 0,
        ...sx as object,
      }}
    >
      <LucideIcon size={size} />
    </Box>
  );
}
