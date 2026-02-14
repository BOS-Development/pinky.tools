import { Character } from "@industry-tool/client/data/models";
import Item from "./item";
import Navbar from "@industry-tool/components/Navbar";
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import AddIcon from '@mui/icons-material/Add';
import RefreshIcon from '@mui/icons-material/Refresh';

export type CharacterListProps = {
  characters: Character[];
};

export default function List(props: CharacterListProps) {
  if (props.characters.length == 0) {
    return (
      <>
        <Navbar />
        <Container maxWidth="lg" sx={{ mt: 4 }}>
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: '60vh',
              textAlign: 'center',
            }}
          >
            <Typography variant="h4" gutterBottom>
              No Characters
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
              Get started by adding your first character
            </Typography>
            <Button
              variant="contained"
              size="large"
              startIcon={<AddIcon />}
              href="api/characters/add"
            >
              Add Character
            </Button>
          </Box>
        </Container>
      </>
    );
  }

  return (
    <>
      <Navbar />
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ mb: 4 }}>
          <Typography variant="h4" gutterBottom>
            Characters
          </Typography>
          <Stack direction="row" spacing={2}>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              href="api/characters/add"
            >
              Add Character
            </Button>
            <Button
              variant="outlined"
              startIcon={<RefreshIcon />}
              href="api/characters/refreshAssets"
            >
              Refresh Assets
            </Button>
          </Stack>
        </Box>
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {
              xs: '1fr',
              sm: 'repeat(2, 1fr)',
              md: 'repeat(3, 1fr)',
            },
            gap: 3,
          }}
        >
          {props.characters.map((char) => {
            return <Item character={char} key={char.id} />;
          })}
        </Box>
      </Container>
    </>
  );
}
