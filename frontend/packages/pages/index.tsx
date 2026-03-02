import { useState, useEffect } from 'react';
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import { FONT_NUMERIC } from "@industry-tool/utils/formatting";

export default function Home() {
  const { data: session, status } = useSession();
  const [assetMetrics, setAssetMetrics] = useState({ totalValue: 0, totalDeficit: 0, activeJobs: 0 });

  useEffect(() => {
    if (status === "authenticated" && session?.providerAccountId) {
      fetch('/api/assets/summary')
        .then(response => {
          if (response.ok) {
            return response.json();
          }
          return { totalValue: 0, totalDeficit: 0, activeJobs: 0 };
        })
        .then(data => setAssetMetrics(data))
        .catch(error => {
          console.error('[Landing] Failed to fetch asset metrics:', error);
        });
    }
  }, [status, session]);

  const isAuthenticated = status === "authenticated";

  return (
    <>
      <Navbar />

      <Box sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'space-between',
        height: '100vh',
        mt: '-64px',
        pt: '64px',
        background: '#0a0e1a',
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
                width: 100,
                height: 'auto',
                mb: 2,
                borderRadius: 2,
                filter: 'drop-shadow(0 0 24px rgba(0, 212, 255, 0.3))'
              }}
            />
            <Typography variant="h1" sx={{ color: '#f1f5f9', mb: 1 }}>
              EVE Industry Tool
            </Typography>
            <Typography variant="body1" sx={{ color: '#94a3b8', mb: 4, maxWidth: 480, mx: 'auto' }}>
              Real-time asset tracking, stockpile management, and market intelligence
              for EVE Online industrialists
            </Typography>

            {isAuthenticated ? (
              <>
                {/* Live Metric Strip */}
                <Box sx={{
                  display: 'flex',
                  justifyContent: 'center',
                  gap: 2,
                  mb: 4,
                  flexWrap: 'wrap',
                }}>
                  <MetricCard
                    label="Asset Value"
                    value={assetMetrics.totalValue === 0
                      ? '—'
                      : `${assetMetrics.totalValue.toLocaleString(undefined, { maximumFractionDigits: 0 })} ISK`
                    }
                    color="#00d4ff"
                  />
                  <MetricCard
                    label="Stockpile Deficit"
                    value={assetMetrics.totalDeficit > 0
                      ? `${assetMetrics.totalDeficit.toLocaleString(undefined, { maximumFractionDigits: 0 })} ISK`
                      : 'None'
                    }
                    color={assetMetrics.totalDeficit > 0 ? '#f43f5e' : '#2dd4bf'}
                  />
                  <MetricCard
                    label="Active Jobs"
                    value={assetMetrics.activeJobs > 0 ? String(assetMetrics.activeJobs) : '—'}
                    color="#fbbf24"
                  />
                </Box>

                <Button
                  variant="contained"
                  size="large"
                  href="/inventory"
                  sx={{
                    px: 4,
                    py: 1.5,
                  }}
                >
                  Open Dashboard
                </Button>
              </>
            ) : (
              <Button
                variant="contained"
                size="large"
                href="/api/auth/login"
                sx={{
                  px: 4,
                  py: 1.5,
                }}
              >
                Sign In with EVE Online
              </Button>
            )}
          </Container>
        </Box>

        {/* Footer */}
        <Box sx={{ py: 2, textAlign: 'center' }}>
          <Typography variant="caption" sx={{ color: 'rgba(148, 163, 184, 0.6)' }}>
            EVE Industry Tool is not affiliated with CCP Games. EVE Online and the EVE logo
            are the intellectual property of CCP hf.
          </Typography>
        </Box>
      </Box>
    </>
  );
}

function MetricCard({ label, value, color }: { label: string; value: string; color: string }) {
  return (
    <Card sx={{
      px: 3,
      py: 2,
      background: '#12151f',
      border: '1px solid rgba(0, 212, 255, 0.08)',
      minWidth: 180,
    }}>
      <Typography variant="caption" sx={{ color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em', fontWeight: 600 }}>
        {label}
      </Typography>
      <Typography variant="h5" sx={{ color, fontWeight: 700, fontFamily: FONT_NUMERIC, mt: 0.5 }}>
        {value}
      </Typography>
    </Card>
  );
}
