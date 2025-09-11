export type MakeTypeFieldsRequired<T, K extends keyof T> = T & Required<Pick<T, K>>;

