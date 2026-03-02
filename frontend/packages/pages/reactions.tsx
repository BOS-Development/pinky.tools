import { AssetsResponse, PlanResponse, PlanSelection, ReactionsResponse, ReactionSystem } from "@industry-tool/client/data/models";
import Navbar from "@industry-tool/components/Navbar";
import PlanSummary from "@industry-tool/components/reactions/PlanSummary";
import ReactionPicker from "@industry-tool/components/reactions/ReactionPicker";
import SettingsToolbar from "@industry-tool/components/reactions/SettingsToolbar";
import ShoppingList from "@industry-tool/components/reactions/ShoppingList";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Loader2 } from "lucide-react";
import { useSession } from 'next-auth/react';
import { useCallback, useEffect, useRef, useState } from 'react';

export type ReactionSettings = {
    system_id: number;
    structure: string;
    rig: string;
    security: string;
    reactions_skill: number;
    facility_tax: number;
    cycle_days: number;
    broker_fee: number;
    sales_tax: number;
    shipping_m3: number;
    shipping_collateral: number;
    input_price: string;
    output_price: string;
    contract_sales: boolean;
    ship_inputs: boolean;
    ship_outputs: boolean;
};

const DEFAULT_SETTINGS: ReactionSettings = {
    system_id: 0,
    structure: 'tatara',
    rig: 't2',
    security: 'null',
    reactions_skill: 5,
    facility_tax: 0.25,
    cycle_days: 7,
    broker_fee: 3.5,
    sales_tax: 2.25,
    shipping_m3: 0,
    shipping_collateral: 0,
    input_price: 'sell',
    output_price: 'sell',
    contract_sales: false,
    ship_inputs: true,
    ship_outputs: true,
};

const TAB_VALUES = ['pick', 'shopping', 'summary'] as const;

export default function Reactions() {
    const { status } = useSession();
    const [tabValue, setTabValue] = useState<string>('pick');
    const [settings, setSettings] = useState<ReactionSettings>(DEFAULT_SETTINGS);
    const [selections, setSelections] = useState<Record<number, number>>({});
    const [systems, setSystems] = useState<ReactionSystem[]>([]);
    const [reactionsData, setReactionsData] = useState<ReactionsResponse | null>(null);
    const [planData, setPlanData] = useState<PlanResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [planLoading, setPlanLoading] = useState(false);
    const [assets, setAssets] = useState<AssetsResponse | null>(null);
    const [stockpileLocationId, setStockpileLocationId] = useState(0);

    // Load persisted state from localStorage after hydration
    const initialized = useRef(false);
    useEffect(() => {
        try {
            const savedTab = localStorage.getItem('reactionsTab');
            if (savedTab) {
                // Migrate old numeric tab index to new string values
                const idx = parseInt(savedTab, 10);
                if (!isNaN(idx) && idx >= 0 && idx < TAB_VALUES.length) {
                    setTabValue(TAB_VALUES[idx]);
                } else if (TAB_VALUES.includes(savedTab as typeof TAB_VALUES[number])) {
                    setTabValue(savedTab);
                }
            }

            const savedSettings = localStorage.getItem('reactionSettings');
            if (savedSettings) setSettings(prev => ({ ...prev, ...JSON.parse(savedSettings) }));

            const savedSelections = localStorage.getItem('reactionSelections');
            if (savedSelections) setSelections(JSON.parse(savedSelections));

            const savedLocation = localStorage.getItem('reactionStockpileLocationId');
            if (savedLocation) setStockpileLocationId(parseInt(savedLocation, 10));
        } catch { }
        initialized.current = true;
    }, []);

    // Persist state to localStorage (skip initial render to avoid overwriting with defaults)
    useEffect(() => {
        if (!initialized.current) return;
        localStorage.setItem('reactionsTab', tabValue);
        localStorage.setItem('reactionSettings', JSON.stringify(settings));
        localStorage.setItem('reactionSelections', JSON.stringify(selections));
        localStorage.setItem('reactionStockpileLocationId', stockpileLocationId.toString());
    }, [tabValue, settings, selections, stockpileLocationId]);

    // Fetch assets when authenticated
    useEffect(() => {
        if (status !== 'authenticated') return;
        fetch('/api/assets/get')
            .then(res => {
                if (res.ok) return res.json();
                throw new Error('Failed to fetch assets');
            })
            .then(data => setAssets(data))
            .catch(err => console.error('Failed to fetch assets:', err));
    }, [status]);

    // Fetch systems once
    useEffect(() => {
        fetch('/api/reactions/systems')
            .then(res => res.json())
            .then(data => setSystems(data))
            .catch(err => console.error('Failed to fetch systems:', err));
    }, []);

    // Fetch reactions when settings change
    const fetchReactions = useCallback(async () => {
        setLoading(true);
        try {
            const params = new URLSearchParams();
            if (settings.system_id > 0) params.set('system_id', settings.system_id.toString());
            params.set('structure', settings.structure);
            params.set('rig', settings.rig);
            params.set('security', settings.security);
            params.set('reactions_skill', settings.reactions_skill.toString());
            params.set('facility_tax', settings.facility_tax.toString());
            params.set('cycle_days', settings.cycle_days.toString());
            params.set('broker_fee', settings.contract_sales ? '0' : settings.broker_fee.toString());
            params.set('sales_tax', settings.contract_sales ? '0' : settings.sales_tax.toString());
            params.set('shipping_m3', settings.shipping_m3.toString());
            params.set('shipping_collateral', settings.shipping_collateral.toString());
            params.set('input_price', settings.input_price);
            params.set('output_price', settings.output_price);
            params.set('ship_inputs', settings.ship_inputs ? '1' : '0');
            params.set('ship_outputs', settings.ship_outputs ? '1' : '0');

            const res = await fetch(`/api/reactions?${params.toString()}`);
            const data = await res.json();
            setReactionsData(data);
        } catch (err) {
            console.error('Failed to fetch reactions:', err);
        } finally {
            setLoading(false);
        }
    }, [settings]);

    useEffect(() => {
        fetchReactions();
    }, [fetchReactions]);

    // Compute plan when selections change and we have reactions data
    const fetchPlan = useCallback(async () => {
        const planSelections: PlanSelection[] = Object.entries(selections)
            .filter(([, count]) => count > 0)
            .map(([id, count]) => ({ reaction_type_id: parseInt(id), instances: count }));

        if (planSelections.length === 0) {
            setPlanData(null);
            return;
        }

        setPlanLoading(true);
        try {
            const res = await fetch('/api/reactions/plan', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    selections: planSelections,
                    ...settings,
                    broker_fee: settings.contract_sales ? 0 : settings.broker_fee,
                    sales_tax: settings.contract_sales ? 0 : settings.sales_tax,
                }),
            });
            const data = await res.json();
            setPlanData(data);
        } catch (err) {
            console.error('Failed to compute plan:', err);
        } finally {
            setPlanLoading(false);
        }
    }, [selections, settings]);

    useEffect(() => {
        fetchPlan();
    }, [fetchPlan]);

    const updateSetting = <K extends keyof ReactionSettings>(key: K, value: ReactionSettings[K]) => {
        setSettings(prev => ({ ...prev, [key]: value }));
    };

    const updateSelection = (reactionTypeId: number, instances: number) => {
        setSelections(prev => {
            const next = { ...prev };
            if (instances <= 0) {
                delete next[reactionTypeId];
            } else {
                next[reactionTypeId] = instances;
            }
            return next;
        });
    };

    const selectedCount = Object.values(selections).filter(v => v > 0).length;

    return (
        <>
            <Navbar />
            <div className="w-full px-4 py-2">
                <SettingsToolbar
                    settings={settings}
                    systems={systems}
                    onSettingChange={updateSetting}
                />

                <Tabs value={tabValue} onValueChange={setTabValue}>
                    <div className="border-b border-[var(--color-border-dim)] mb-4">
                        <TabsList className="bg-transparent">
                            <TabsTrigger value="pick">Pick Reactions</TabsTrigger>
                            <TabsTrigger value="shopping">
                                Shopping List{selectedCount > 0 ? ` (${selectedCount})` : ''}
                            </TabsTrigger>
                            <TabsTrigger value="summary">Plan Summary</TabsTrigger>
                        </TabsList>
                    </div>

                    {loading ? (
                        <div className="flex justify-center py-16">
                            <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary-cyan)]" />
                        </div>
                    ) : !reactionsData ? (
                        <p className="py-8 text-center text-[var(--color-text-secondary)]">
                            No reaction data available. Ensure SDE data has been imported.
                        </p>
                    ) : (
                        <>
                            <TabsContent value="pick">
                                <ReactionPicker
                                    reactions={reactionsData.reactions}
                                    meFactor={reactionsData.me_factor}
                                    selections={selections}
                                    onSelectionChange={updateSelection}
                                />
                            </TabsContent>
                            <TabsContent value="shopping">
                                <ShoppingList
                                    planData={planData}
                                    loading={planLoading}
                                    assets={assets}
                                    isAuthenticated={status === 'authenticated'}
                                    stockpileLocationId={stockpileLocationId}
                                    onStockpileLocationChange={setStockpileLocationId}
                                />
                            </TabsContent>
                            <TabsContent value="summary">
                                <PlanSummary
                                    planData={planData}
                                    reactionsData={reactionsData}
                                    selections={selections}
                                    loading={planLoading}
                                />
                            </TabsContent>
                        </>
                    )}
                </Tabs>
            </div>
        </>
    );
}
