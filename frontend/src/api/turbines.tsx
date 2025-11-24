import { createSlice } from "@reduxjs/toolkit"
import { fetchFromAPI } from "./fetch"
import type { Turbine } from "./interfaces"
import { turbinesMock } from "./mock"
import { useSelector } from "react-redux"


interface TurbinesState {
    filter: string;
    items: Turbine[];
}

const turbinesSlice = createSlice({
    name: "turbines",
    initialState: {
        filter: "",
        items: [],
    } as TurbinesState,
    reducers: {
        setTurbinesFilter: (state, action) => {
            state.filter = action.payload.trim();
        }
    },
})

export const useTurbinesFilter = () => useSelector((state: { turbines: TurbinesState }) => state.turbines.filter);

export const { setTurbinesFilter } = turbinesSlice.actions;

export const turbinesReducer = turbinesSlice.reducer;

export const fetchTurbines = (titleFilter: string) => fetchFromAPI<Turbine[]>(titleFilter
        ? `/turbines/?turbineTitle=${encodeURIComponent(titleFilter)}`
        : `/turbines/`,
        turbinesMock.filter(turbine => turbine.title.toLowerCase().includes(titleFilter.toLowerCase()))
    );

export const fetchTurbine = async (id: number) => fetchFromAPI<Turbine | null>(
    `/turbines/${id}`,
    turbinesMock.find(turbine => turbine.id === id) || null
)