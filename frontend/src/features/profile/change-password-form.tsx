import { FormEvent, useState } from 'react';
import { AxiosError } from 'axios';
import { changeUserPassword } from '@/api/users';
import { useAuthStore } from '@/app/store/auth-store';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

export const ChangePasswordForm = () => {
  const user = useAuthStore((s) => s.user);
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setStatus(null);
    setError(null);

    if (!user?.id) {
      setError('Не найден user id в сессии. Перелогинься.');
      return;
    }

    try {
      await changeUserPassword(user.id, { old_password: oldPassword, new_password: newPassword });
      setStatus('Пароль успешно изменён.');
      setOldPassword('');
      setNewPassword('');
    } catch (e) {
      setError(parseError(e));
    }
  };

  return (
    <form className="theme-card space-y-3 rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm" onSubmit={onSubmit}>
      <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Сменить пароль</h2>
      <input className="form-input w-full rounded-lg px-3 py-2" type="password" placeholder="Старый пароль" value={oldPassword} onChange={(e) => setOldPassword(e.target.value)} required />
      <input className="form-input w-full rounded-lg px-3 py-2" type="password" placeholder="Новый пароль" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} required />
      {status && <p className="text-sm text-emerald-700">{status}</p>}
      {error && <p className="field-error">{error}</p>}
      <button className="interactive-chip theme-button px-4 py-2">Обновить пароль</button>
    </form>
  );
};
