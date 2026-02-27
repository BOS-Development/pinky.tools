import { Corporation } from "@industry-tool/client/data/models";
import { corporationScopesUpToDate } from "@industry-tool/client/scopes";
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardMedia from '@mui/material/CardMedia';
import Typography from '@mui/material/Typography';
import BusinessIcon from '@mui/icons-material/Business';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Tooltip from '@mui/material/Tooltip';
import Button from '@mui/material/Button';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';

export type CorporationItemProps = {
  corporation: Corporation;
};

export default function Item(props: CorporationItemProps) {
  const needsUpdate = !corporationScopesUpToDate(props.corporation);

  return (
    <Card
      sx={{
        maxWidth: 345,
        transition: 'transform 0.2s, box-shadow 0.2s',
        '&:hover': {
          transform: 'translateY(-4px)',
          boxShadow: 6,
        },
        ...(needsUpdate && {
          border: '2px solid',
          borderColor: 'warning.main',
        }),
      }}
    >
      <Box
        sx={{
          height: 200,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'linear-gradient(135deg, #1976d2 0%, #42a5f5 100%)',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        <CardMedia
          component="img"
          height="200"
          image={`https://images.evetech.net/corporations/${props.corporation.id}/logo?size=256&tenant=tranquility`}
          alt={props.corporation.name}
          sx={{
            objectFit: 'contain',
            padding: 2,
            filter: 'drop-shadow(0 4px 8px rgba(0,0,0,0.3))',
          }}
          onError={(e) => {
            const target = e.target as HTMLImageElement;
            target.style.display = 'none';
          }}
        />
        {needsUpdate && (
          <Tooltip title="Scopes need updating">
            <WarningAmberIcon
              color="warning"
              sx={{
                position: 'absolute',
                top: 8,
                right: 8,
                fontSize: 32,
                filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.5))',
              }}
            />
          </Tooltip>
        )}
      </Box>
      <CardContent>
        <Box sx={{ mb: 1 }}>
          <Chip
            icon={<BusinessIcon />}
            label="Corporation"
            size="small"
            color="primary"
            variant="outlined"
            sx={{ mb: 1 }}
          />
        </Box>
        <Typography
          variant="h5"
          component="div"
          sx={{
            fontWeight: 600,
            color: 'primary.light',
          }}
        >
          {props.corporation.name}
        </Typography>
        {needsUpdate && (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 1 }}>
            <Tooltip title="This corporation needs to be re-authorized to grant new permissions">
              <WarningAmberIcon color="warning" />
            </Tooltip>
            <Button
              size="small"
              variant="outlined"
              color="warning"
              href="/api/corporations/add"
            >
              Re-authorize
            </Button>
          </Box>
        )}
      </CardContent>
    </Card>
  );
}
