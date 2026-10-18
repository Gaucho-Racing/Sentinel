export function formatPhoneNumber(phone: string): string {
  const us = /^\+1(\d{3})(\d{3})(\d{4})$/.exec(phone)
  return us ? `+1 (${us[1]}) ${us[2]}-${us[3]}` : phone
}
