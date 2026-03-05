import Head from "next/head";
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
      <Head><title>pinky.tools</title></Head>
      <Navbar />

      <div className="flex flex-col items-center justify-between min-h-screen -mt-16 pt-16 bg-background-void text-center px-3 pb-4">
        <div className="flex-1 flex flex-col justify-center w-full">
          <div className="max-w-2xl mx-auto">
            <img
              src="https://images.evetech.net/types/23773/render?size=1024"
              alt="Ragnarok Titan"
              className="w-[100px] h-auto mb-2 rounded shadow-glow-lg"
            />
            <h1 className="text-3xl font-bold tracking-tight text-text-emphasis mb-1">
              pinky.tools
            </h1>
            <p className="text-[0.9375rem] leading-relaxed text-text-secondary mb-4 max-w-[480px] mx-auto">
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
                      ? '0 ISK'
                      : `${assetMetrics.totalValue.toLocaleString(undefined, { maximumFractionDigits: 0 })} ISK`
                    }
                    color="var(--color-data-value)"
                  />
                  <MetricCard
                    label="Stockpile Deficit"
                    value={assetMetrics.totalDeficit > 0
                      ? `${assetMetrics.totalDeficit.toLocaleString(undefined, { maximumFractionDigits: 0 })} ISK`
                      : '0 ISK'
                    }
                    color={assetMetrics.totalDeficit > 0 ? 'var(--color-danger-rose)' : 'var(--color-success-teal)'}
                  />
                  <MetricCard
                    label="Active Jobs"
                    value={assetMetrics.activeJobs > 0 ? String(assetMetrics.activeJobs) : '0'}
                    color="var(--color-manufacturing-amber)"
                  />
                </div>
              </>
            ) : (
              <a
                href="/api/auth/login"
                className="inline-block px-8 py-3 bg-primary text-background-void font-medium text-sm rounded-sm shadow-glow-sm hover:shadow-glow-md transition-shadow"
              >
                Sign In with EVE Online
              </a>
            )}
          </div>
        </div>

        {/* Quick Access */}
        {isAuthenticated && (
          <div className="w-full max-w-3xl mx-auto pb-8">
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <QuickLink href="/inventory" label="Inventory" description="View all assets" icon="/icons/inventory-empty.png" />
              <QuickLink href="/stockpiles" label="Stockpiles" description="Track targets" icon="/icons/isk.png" />
              <QuickLink href="/industry" label="Industry" description="Manage jobs" icon="/icons/industry.png" />
              <QuickLink href="/marketplace" label="Marketplace" description="Browse listings" icon="/icons/trading.png" />
              <QuickLink href="/reactions" label="Reactions" description="Calculator" icon="/icons/reactions.png" />
              <QuickLink href="/production-plans" label="Plans" description="Production plans" icon="/icons/plans.png" />
              <QuickLink href="/contacts" label="Contacts" description="Trading network" icon="/icons/contacts.png" />
              <QuickLink href="/hauling" label="Hauling" description="Run logistics" icon="/icons/logistics.png" />
            </div>
          </div>
        )}

        {/* Footer */}
        <div className="py-2 text-center">
          <p className="text-xs text-text-muted">
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
    <div className="px-3 py-2 bg-background-panel border border-border-dim rounded-sm min-w-[180px]">
      <p className="text-xs text-text-muted uppercase tracking-wider font-semibold">
        {label}
      </p>
      <p className="text-lg font-bold mt-0.5" style={{ color, fontFamily: FONT_NUMERIC }}>
        {value}
      </p>
    </div>
  );
}

function QuickLink({ href, label, description, icon }: { href: string; label: string; description: string; icon?: string }) {
  return (
    <a
      href={href}
      className="px-3 py-3 bg-background-panel border border-border-dim rounded-sm hover:border-primary/40 hover:shadow-glow-sm transition-all group"
    >
      {icon && <img src={icon} alt="" className="h-8 w-8 object-contain mb-1.5 opacity-85" />}
      <p className="text-sm font-semibold text-text-emphasis group-hover:text-primary transition-colors">
        {label}
      </p>
      <p className="text-xs text-text-muted mt-0.5">
        {description}
      </p>
    </a>
  );
}
