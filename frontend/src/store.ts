import { combineReducers, configureStore } from "@reduxjs/toolkit"
import { turbinesReducer } from "./api/turbines"


export default configureStore({
    reducer: combineReducers({
        turbines: turbinesReducer
    })
})