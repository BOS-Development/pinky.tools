import { Character } from "@industry-tool/client/data/models";
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardMedia from '@mui/material/CardMedia';
import Typography from '@mui/material/Typography';

export type CharacterItemProps = {
  character: Character;
};

export default function Item(props: CharacterItemProps) {
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
      <CardContent>
        <Typography gutterBottom variant="h6" component="div">
          {props.character.name}
        </Typography>
      </CardContent>
    </Card>
  );
}
