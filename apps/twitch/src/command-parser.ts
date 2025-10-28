export interface ModifyStatParseResult {
  mentionedLogin: string;
  statColumn: string;
  amount: number;
  remove: boolean;
  error?: string;
}

/** Parse commands: !addstat @user stat amount OR !rmstat @user stat amount */
export const parseModifyStatCommand = (
  text: string,
  isRemove: boolean
): ModifyStatParseResult => {
  const parts = text.trim().split(/\s+/);
  if (parts.length < 4) {
    return {
      mentionedLogin: "",
      statColumn: "",
      amount: 0,
      remove: isRemove,
      error: "Expected format: !addstat @user stat amount",
    };
  }

  const mention = parts.find((p) => p.startsWith("@"));
  if (!mention) {
    return {
      mentionedLogin: "",
      statColumn: "",
      amount: 0,
      remove: isRemove,
      error: "No user mention found",
    };
  }

  const mentionedLogin = mention.slice(1).toLowerCase();
  const statColumn = parts[2];
  const amountStr = parts[3];

  if (!statColumn || !amountStr) {
    return {
      mentionedLogin,
      statColumn: "",
      amount: 0,
      remove: isRemove,
      error: "Expected 'stat amount'",
    };
  }

  const amount = parseInt(amountStr, 10);
  if (Number.isNaN(amount)) {
    return {
      mentionedLogin,
      statColumn,
      amount: 0,
      remove: isRemove,
      error: "Invalid number",
    };
  }

  return {
    mentionedLogin,
    statColumn,
    amount,
    remove: isRemove,
  };
};
