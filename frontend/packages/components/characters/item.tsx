import { Character } from "@industry-tool/client/data/models";
import { characterScopesUpToDate } from "@industry-tool/client/scopes";
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardMedia from '@mui/material/CardMedia';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Tooltip from '@mui/material/Tooltip';
import Button from '@mui/material/Button';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';

export type CharacterItemProps = {
  character: Character;
};

export default function Item(props: CharacterItemProps) {
  const needsUpdate = !characterScopesUpToDate(props.character);

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
      <Box sx={{ position: 'relative' }}>
        <CardMedia
          component="img"
          image={`https://image.eveonline.com/Character/${props.character.id}_128.jpg`}
          alt={props.character.name}
          sx={{
            width: '100%',
            maxWidth: 192,
            aspectRatio: '1',
            objectFit: 'contain',
            backgroundColor: '#1a1a1a',
            margin: '0 auto'
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
        <Typography gutterBottom variant="h6" component="div">
          {props.character.name}
        </Typography>
        {needsUpdate && (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 1 }}>
            <Tooltip title="This character needs to be re-authorized to grant new permissions">
              <WarningAmberIcon color="warning" />
            </Tooltip>
            <Button
              size="small"
              variant="outlined"
              color="warning"
              href="/api/characters/add"
            >
              Re-authorize
            </Button>
          </Box>
        )}
      </CardContent>
    </Card>
  );
}
