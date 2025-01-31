export type codeground = {
  id: string;
  name: String;
  codegroundType: codegroundType;
  createdAt: Date;
};

enum codegroundType {
  REACT,
  NODE,
}
