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
import ErrorIcon from '@mui/icons-material/Error';

export type CharacterItemProps = {
  character: Character;
};

export default function Item(props: CharacterItemProps) {
  const needsReauth = props.character.needsReauth === true;
  const needsScopeUpdate = !needsReauth && !characterScopesUpToDate(props.character);

  return (
    <Card
      sx={{
        maxWidth: 345,
        transition: 'transform 0.2s, box-shadow 0.2s',
        '&:hover': {
          transform: 'translateY(-4px)',
          boxShadow: 6,
        },
        ...((needsReauth) && {
          border: '2px solid',
          borderColor: 'error.main',
        }),
        ...((!needsReauth && needsScopeUpdate) && {
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
        {needsReauth && (
          <Tooltip title="Authorization revoked — re-authorize required">
            <ErrorIcon
              color="error"
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
        {needsScopeUpdate && (
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
        {needsReauth && (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 1 }}>
            <Tooltip title="This character's ESI authorization has been revoked and requires re-authorization">
              <ErrorIcon color="error" />
            </Tooltip>
            <Button
              size="small"
              variant="outlined"
              color="error"
              href="/api/characters/add"
            >
              Re-authorize
            </Button>
          </Box>
        )}
        {needsScopeUpdate && (
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
