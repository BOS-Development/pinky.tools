import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { ReactionSettings } from "@industry-tool/pages/reactions";

type Props = {
  open: boolean;
  onClose: () => void;
  settings: ReactionSettings;
  onSettingChange: <K extends keyof ReactionSettings>(key: K, value: ReactionSettings[K]) => void;
};

export default function SettingsModal({ open, onClose, settings, onSettingChange }: Props) {
  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose(); }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Advanced Settings</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-4 pt-1">
          <p className="text-sm font-medium text-[var(--color-text-secondary)]">Skills & Fees</p>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[var(--color-text-secondary)]">Reactions Skill Level</label>
            <Input
              type="number"
              value={settings.reactions_skill}
              onChange={(e) => {
                const val = parseInt(e.target.value);
                if (val >= 0 && val <= 5) onSettingChange('reactions_skill', val);
              }}
              min={0}
              max={5}
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[var(--color-text-secondary)]">Facility Tax (%)</label>
            <Input
              type="number"
              value={settings.facility_tax}
              onChange={(e) => onSettingChange('facility_tax', parseFloat(e.target.value) || 0)}
              min={0}
              step={0.25}
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[var(--color-text-secondary)]">Broker Fee (%)</label>
            <Input
              type="number"
              value={settings.broker_fee}
              onChange={(e) => onSettingChange('broker_fee', parseFloat(e.target.value) || 0)}
              min={0}
              step={0.1}
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[var(--color-text-secondary)]">Sales Tax (%)</label>
            <Input
              type="number"
              value={settings.sales_tax}
              onChange={(e) => onSettingChange('sales_tax', parseFloat(e.target.value) || 0)}
              min={0}
              step={0.1}
            />
          </div>

          <p className="text-sm font-medium text-[var(--color-text-secondary)] mt-2">Pricing</p>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[var(--color-text-secondary)]">Input Price</label>
            <Select value={settings.input_price} onValueChange={(val) => onSettingChange('input_price', val)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="sell">Jita Sell</SelectItem>
                <SelectItem value="buy">Jita Buy</SelectItem>
                <SelectItem value="split">Split (avg)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[var(--color-text-secondary)]">Output Price</label>
            <Select value={settings.output_price} onValueChange={(val) => onSettingChange('output_price', val)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="sell">Jita Sell</SelectItem>
                <SelectItem value="buy">Jita Buy</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center justify-between">
            <label className="text-sm text-[var(--color-text-primary)]">
              Contract Sales (no broker fee / sales tax)
            </label>
            <Switch
              checked={settings.contract_sales}
              onCheckedChange={(checked) => onSettingChange('contract_sales', checked)}
            />
          </div>

          <p className="text-sm font-medium text-[var(--color-text-secondary)] mt-2">Shipping</p>

          <div className="flex items-center justify-between">
            <label className="text-sm text-[var(--color-text-primary)]">Ship Inputs</label>
            <Switch
              checked={settings.ship_inputs}
              onCheckedChange={(checked) => onSettingChange('ship_inputs', checked)}
            />
          </div>

          <div className="flex items-center justify-between">
            <label className="text-sm text-[var(--color-text-primary)]">Ship Outputs</label>
            <Switch
              checked={settings.ship_outputs}
              onCheckedChange={(checked) => onSettingChange('ship_outputs', checked)}
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[var(--color-text-secondary)]">Shipping ISK/m3</label>
            <Input
              type="number"
              value={settings.shipping_m3}
              onChange={(e) => onSettingChange('shipping_m3', parseFloat(e.target.value) || 0)}
              min={0}
              step={100}
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[var(--color-text-secondary)]">Collateral (%)</label>
            <Input
              type="number"
              value={settings.shipping_collateral * 100}
              onChange={(e) => onSettingChange('shipping_collateral', (parseFloat(e.target.value) || 0) / 100)}
              min={0}
              max={100}
              step={0.5}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
