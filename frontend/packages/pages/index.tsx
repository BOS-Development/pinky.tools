import { useState, useEffect } from 'react';
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
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

      <div className="flex flex-col items-center justify-between h-screen -mt-16 pt-16 bg-[#0a0e1a] text-center px-3 pb-4">
        <div className="flex-1 flex flex-col justify-center w-full">
          <div className="max-w-2xl mx-auto">
            <img
              src="https://images.evetech.net/types/23773/render?size=512"
              alt="Ragnarok Titan"
              className="w-[100px] h-auto mb-2 rounded drop-shadow-[0_0_24px_rgba(0,212,255,0.3)]"
            />
            <h1 className="text-3xl font-bold tracking-tight text-[#f1f5f9] mb-1">
              pinky.tools
            </h1>
            <p className="text-[0.9375rem] leading-relaxed text-[#94a3b8] mb-4 max-w-[480px] mx-auto">
              Real-time asset tracking, stockpile management, and market intelligence
              for EVE Online industrialists
            </p>

            {isAuthenticated ? (
              <>
                {/* Live Metric Strip */}
                <div className="flex justify-center gap-2 mb-4 flex-wrap">
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
                </div>

                <a
                  href="/inventory"
                  className="inline-block px-8 py-3 bg-[#00d4ff] text-[#0a0a0f] font-medium text-sm rounded-sm shadow-[0_0_8px_rgba(0,212,255,0.25)] hover:shadow-[0_0_12px_rgba(0,212,255,0.35)] transition-shadow"
                >
                  Open Dashboard
                </a>
              </>
            ) : (
              <a
                href="/api/auth/login"
                className="inline-block px-8 py-3 bg-[#00d4ff] text-[#0a0a0f] font-medium text-sm rounded-sm shadow-[0_0_8px_rgba(0,212,255,0.25)] hover:shadow-[0_0_12px_rgba(0,212,255,0.35)] transition-shadow"
              >
                Sign In with EVE Online
              </a>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="py-2 text-center">
          <p className="text-xs text-[rgba(148,163,184,0.6)]">
            pinky.tools is not affiliated with CCP Games. EVE Online and the EVE logo
            are the intellectual property of CCP hf.
          </p>
        </div>
      </div>
    </>
  );
}

function MetricCard({ label, value, color }: { label: string; value: string; color: string }) {
  return (
    <div className="px-3 py-2 bg-[#12151f] border border-[rgba(0,212,255,0.08)] rounded-sm min-w-[180px]">
      <p className="text-xs text-[#64748b] uppercase tracking-wider font-semibold">
        {label}
      </p>
      <p className="text-lg font-bold mt-0.5" style={{ color, fontFamily: FONT_NUMERIC }}>
        {value}
      </p>
    </div>
  );
}
