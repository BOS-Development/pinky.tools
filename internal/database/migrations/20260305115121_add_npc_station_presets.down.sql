-- Migration: add_npc_station_presets
-- Created: Thu Mar  5 11:51:21 AM PST 2026

delete from trading_stations
where station_id in (60008494, 60011866, 60004588, 60005686);
