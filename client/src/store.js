import { configureStore, createSlice } from "@reduxjs/toolkit";

const userInfo = createSlice({
  name: "myInfoReducer",
  initialState: {
    isLogin: false,
    needValidation: true,
    isAdmin: false,
    id: "",
    name: "",
    money: 0,
  },
  reducers: {
    update: (state, action) => {
      state = action.payload;
      return state;
    },
  },
});

export const { update } = userInfo.actions;
export default configureStore({ reducer: userInfo.reducer });
