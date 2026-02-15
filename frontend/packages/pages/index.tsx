import { getServerSession } from "next-auth/next";
import { authOptions } from "../../pages/api/auth/[...nextauth]";
import Navbar from "@industry-tool/components/Navbar";
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import Stack from '@mui/material/Stack';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';

export default async function Home() {
  const session = await getServerSession(authOptions);
  const isAuthenticated = !!session;

  // Fetch assets summary for authenticated users
  let assetMetrics = { totalValue: 0, totalDeficit: 0 };
  if (isAuthenticated && session?.providerAccountId) {
    try {
      const backend = process.env.BACKEND_URL as string;
      const backendKey = process.env.BACKEND_KEY as string;

      const response = await fetch(`${backend}v1/assets/summary`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'USER-ID': session.providerAccountId,
          'BACKEND-KEY': backendKey,
        },
        cache: 'no-store', // Don't cache to always show fresh data
      });

      if (response.ok) {
        assetMetrics = await response.json();
      }
    } catch (error) {
      console.error('[Landing] Failed to fetch asset metrics:', error);
      // Silently fail and show 0 values rather than breaking the page
    }
  }

  return (
    <>
      <Navbar />

      {/* Hero Section */}
      <Box sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'space-between',
        height: '100vh',
        mt: '-64px',
        pt: '64px',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        textAlign: 'center',
        px: 3,
        pb: 4,
      }}>
        <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', width: '100%' }}>
          <Container maxWidth="md">
            <Box
              component="img"
              src="https://images.evetech.net/types/23773/render?size=512"
              alt="Ragnarok Titan"
              sx={{
                width: 120,
                height: 'auto',
                mb: 2,
                borderRadius: 2,
                filter: 'drop-shadow(0 0 20px rgba(255,255,255,0.3))'
              }}
            />
          <Typography variant="h1" color="white" gutterBottom>
            Master Your EVE Online Assets
          </Typography>
          <Typography variant="h5" color="rgba(255, 255, 255, 0.9)" sx={{ mb: 3 }}>
            Real-time asset tracking, stockpile management, and market intelligence
            for EVE Online industrialists and corporations
          </Typography>

          {isAuthenticated ? (
            <>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} justifyContent="center" sx={{ mb: 4 }}>
                <Button
                  variant="contained"
                  size="large"
                  href="/characters"
                  sx={{
                    fontSize: '1.1rem',
                    px: 4,
                    py: 1.5,
                    backgroundColor: 'white',
                    color: 'primary.main',
                    '&:hover': { backgroundColor: 'rgba(255, 255, 255, 0.9)' }
                  }}
                >
                  Characters
                </Button>
                <Button
                  variant="outlined"
                  size="large"
                  href="/inventory"
                  sx={{
                    fontSize: '1.1rem',
                    px: 4,
                    py: 1.5,
                    borderColor: 'white',
                    color: 'white',
                    '&:hover': {
                      borderColor: 'white',
                      backgroundColor: 'rgba(255, 255, 255, 0.1)'
                    }
                  }}
                >
                  View Assets
                </Button>
                <Button
                  variant="outlined"
                  size="large"
                  href="/stockpiles"
                  sx={{
                    fontSize: '1.1rem',
                    px: 4,
                    py: 1.5,
                    borderColor: 'white',
                    color: 'white',
                    '&:hover': {
                      borderColor: 'white',
                      backgroundColor: 'rgba(255, 255, 255, 0.1)'
                    }
                  }}
                >
                  Manage Stockpiles
                </Button>
              </Stack>

              {/* Metrics Section */}
              <Box sx={{
                display: 'grid',
                gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' },
                gap: 3,
                maxWidth: 'lg',
                mx: 'auto'
              }}>
                {/* Total Asset Value Card */}
                <Card sx={{
                  p: 3,
                  background: 'linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%)',
                  color: 'white',
                  transition: 'transform 0.2s, box-shadow 0.2s',
                  '&:hover': { transform: 'translateY(-4px)', boxShadow: 6 },
                }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <TrendingUpIcon sx={{ fontSize: 48 }} />
                    <Box>
                      <Typography variant="h4" fontWeight={600} gutterBottom>
                        {assetMetrics.totalValue === 0 ? (
                          'Dirty Poor'
                        ) : (
                          `${assetMetrics.totalValue.toLocaleString(undefined, {
                            maximumFractionDigits: 0
                          })} ISK`
                        )}
                      </Typography>
                      <Typography variant="body1" sx={{ opacity: 0.9 }}>
                        Total Asset Value
                      </Typography>
                    </Box>
                  </Box>
                </Card>

                {/* Stockpile Deficit Card */}
                <Card sx={{
                  p: 3,
                  background: assetMetrics.totalDeficit > 0
                    ? 'linear-gradient(135deg, #991b1b 0%, #dc2626 100%)'
                    : 'linear-gradient(135deg, #166534 0%, #22c55e 100%)',
                  color: 'white',
                  transition: 'transform 0.2s, box-shadow 0.2s',
                  '&:hover': { transform: 'translateY(-4px)', boxShadow: 6 },
                }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <WarningAmberIcon sx={{ fontSize: 48 }} />
                    <Box>
                      <Typography variant="h4" fontWeight={600} gutterBottom>
                        {assetMetrics.totalDeficit > 0 ? (
                          `${assetMetrics.totalDeficit.toLocaleString(undefined, {
                            maximumFractionDigits: 0
                          })} ISK`
                        ) : (
                          'No Deficits'
                        )}
                      </Typography>
                      <Typography variant="body1" sx={{ opacity: 0.9 }}>
                        {assetMetrics.totalDeficit > 0 ? 'Stockpile Deficit Cost' : 'All Stockpiles Met'}
                      </Typography>
                    </Box>
                  </Box>
                </Card>
              </Box>
            </>
          ) : (
            <Button
              variant="contained"
              size="large"
              href="/api/auth/signin"
              sx={{
                fontSize: '1.1rem',
                px: 4,
                py: 1.5,
                backgroundColor: 'white',
                color: 'primary.main',
                '&:hover': { backgroundColor: 'rgba(255, 255, 255, 0.9)' }
              }}
            >
              Sign In with EVE Online
            </Button>
          )}
          </Container>
        </Box>

        {/* Footer */}
        <Box sx={{ py: 2, textAlign: 'center' }}>
          <Typography variant="caption" color="rgba(255, 255, 255, 0.7)">
            EVE Industry Tool is not affiliated with CCP Games. EVE Online and the EVE logo
            are the intellectual property of CCP hf.
          </Typography>
        </Box>
      </Box>
    </>
  );
}
