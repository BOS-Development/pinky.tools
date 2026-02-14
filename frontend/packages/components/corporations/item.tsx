import { Corporation } from "@industry-tool/client/data/models";
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardMedia from '@mui/material/CardMedia';
import Typography from '@mui/material/Typography';
import BusinessIcon from '@mui/icons-material/Business';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';

export type CorporationItemProps = {
  corporation: Corporation;
};

export default function Item(props: CorporationItemProps) {
  return (
    <Card
      sx={{
        maxWidth: 345,
        transition: 'transform 0.2s, box-shadow 0.2s',
        '&:hover': {
          transform: 'translateY(-4px)',
          boxShadow: 6,
        }
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
            // Fallback to icon if logo fails to load
            const target = e.target as HTMLImageElement;
            target.style.display = 'none';
          }}
        />
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
      </CardContent>
    </Card>
  );
}
