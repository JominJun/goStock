import { configureStore, createSlice } from "@reduxjs/toolkit";

const myInfo = createSlice({
  name: "myInfoReducer",
  initialState: {
    isLogin: false,
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

export const { update } = myInfo.actions;
export default configureStore({ reducer: myInfo.reducer });
