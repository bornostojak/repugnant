// $rPg: TypeScript request validation, typescript, api
// $~ Validates an input before it reaches the API boundary.
export const valid = (value: string) => value.trim().length > 0
