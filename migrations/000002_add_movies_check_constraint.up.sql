alter table movies add constraint movies_runtime_check check (runtime >= 0);

alter table movies add constraint movies_year_check check (year between 1888 and extract(year from now()));

alter table movies add constraint movies_genres_length_check check (array_length(genres, 1) between 1 and 5);