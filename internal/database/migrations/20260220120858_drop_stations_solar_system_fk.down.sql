alter table stations add constraint stations_solar_system_id_fkey
    foreign key (solar_system_id) references solar_systems(solar_system_id);
