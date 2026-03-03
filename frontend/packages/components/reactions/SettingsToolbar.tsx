import { useState } from 'react';
import { Settings2 } from 'lucide-react';
import { Input } from "@/components/ui/input";
import {
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
} from "@/components/ui/select";
import { Combobox, type ComboboxOption } from "@/components/ui/combobox";
import { ReactionSettings } from "@industry-tool/pages/reactions";
import { ReactionSystem } from "@industry-tool/client/data/models";
import SettingsModal from './SettingsModal';

type Props = {
  settings: ReactionSettings;
  systems: ReactionSystem[];
  onSettingChange: <K extends keyof ReactionSettings>(key: K, value: ReactionSettings[K]) => void;
};

export default function SettingsToolbar({ settings, systems, onSettingChange }: Props) {
  const [modalOpen, setModalOpen] = useState(false);

  const systemOptions: ComboboxOption[] = systems.map(s => ({
    value: s.system_id.toString(),
    label: s.name,
  }));

  return (
    <>
      <div className="flex flex-wrap gap-3 items-center py-3 px-1">
        <Combobox
          options={systemOptions}
          value={settings.system_id ? settings.system_id.toString() : ''}
          onValueChange={(val) => onSettingChange('system_id', val ? parseInt(val) : 0)}
          placeholder="System"
          searchPlaceholder="Search systems..."
          triggerClassName="w-56"
        />

        <Select value={settings.structure} onValueChange={(val) => onSettingChange('structure', val)}>
          <SelectTrigger className="w-32 h-9">
            <SelectValue placeholder="Structure" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="tatara">Tatara</SelectItem>
            <SelectItem value="athanor">Athanor</SelectItem>
          </SelectContent>
        </Select>

        <Select value={settings.rig} onValueChange={(val) => onSettingChange('rig', val)}>
          <SelectTrigger className="w-24 h-9">
            <SelectValue placeholder="Rig" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="t2">T2</SelectItem>
            <SelectItem value="t1">T1</SelectItem>
            <SelectItem value="none">None</SelectItem>
          </SelectContent>
        </Select>

        <Select value={settings.security} onValueChange={(val) => onSettingChange('security', val)}>
          <SelectTrigger className="w-32 h-9">
            <SelectValue placeholder="Security" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="null">Null / WH</SelectItem>
            <SelectItem value="low">Lowsec</SelectItem>
            <SelectItem value="high">Highsec</SelectItem>
          </SelectContent>
        </Select>

        <div className="flex items-center gap-1">
          <label className="text-xs text-[var(--color-text-secondary)]">Cycle</label>
          <Input
            type="number"
            value={settings.cycle_days}
            onChange={(e) => {
              const val = parseInt(e.target.value);
              if (val >= 1 && val <= 30) onSettingChange('cycle_days', val);
            }}
            className="w-16 h-9 text-center"
            min={1}
            max={30}
          />
        </div>

        <button
          onClick={() => setModalOpen(true)}
          className="p-2 rounded-sm text-[var(--color-primary-cyan)] hover:bg-[var(--color-surface-elevated)] transition-colors"
          title="Advanced Settings"
        >
          <Settings2 className="h-5 w-5" />
        </button>
      </div>

      <SettingsModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        settings={settings}
        onSettingChange={onSettingChange}
      />
    </>
  );
}
