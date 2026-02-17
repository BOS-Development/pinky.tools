import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';

export default function Unuathorized() {
  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Paper
        elevation={6}
        sx={{
          p: 4,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          maxWidth: 400,
          textAlign: 'center',
        }}
      >
        <LockOutlinedIcon sx={{ fontSize: 60, mb: 2, color: 'primary.main' }} />
        <Typography variant="h5" gutterBottom>
          Authentication Required
        </Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
          You must sign in to access this page
        </Typography>
        <Button
          variant="contained"
          size="large"
          href="api/auth/login"
          fullWidth
        >
          Sign In
        </Button>
      </Paper>
    </Box>
  );
}
