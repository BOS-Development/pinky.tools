-- Migration: add_npc_station_presets
-- Created: Thu Mar  5 11:51:21 AM PST 2026

insert into trading_stations (station_id, name, system_id, region_id, is_preset)
values
	(60008494, 'Amarr VIII (Oris) - Emperor Family Academy',                30002187, 10000043, true),
	(60011866, 'Dodixie IX - Moon 20 - Federation Navy Assembly Plant',     30002659, 10000032, true),
	(60004588, 'Rens VI - Moon 8 - Brutor Tribe Treasury',                  30002510, 10000030, true),
	(60005686, 'Hek VIII - Moon 12 - Boundless Creation Factory',           30002053, 10000042, true)
on conflict (station_id) do nothing;
