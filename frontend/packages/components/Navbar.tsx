'use client';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import RocketLaunchIcon from '@mui/icons-material/RocketLaunch';

export default function Navbar() {
  return (
    <>
      <AppBar position="fixed">
        <Toolbar>
          <IconButton
            size="large"
            edge="start"
            color="inherit"
            aria-label="menu"
            sx={{ mr: 2 }}
          >
            <RocketLaunchIcon />
          </IconButton>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            EVE Industry Tool
          </Typography>
          <Button color="inherit" href="/characters">
            Characters
          </Button>
          <Button color="inherit" href="/corporations">
            Corporations
          </Button>
          <Button color="inherit" href="/inventory">
            Inventory
          </Button>
          <Button color="inherit" href="/stockpiles">
            Stockpiles
          </Button>
        </Toolbar>
      </AppBar>
      <Toolbar />
    </>
  );
}
