import { useEffect, useState } from 'react';
import { AxiosError } from 'axios';
import { getCurrentUser } from '@/api/auth';
import { useAuthStore } from '@/app/store/auth-store';
import { UserDetails } from '@/api/users';

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

export const ProfileCard = () => {
  const authUser = useAuthStore((s) => s.user);
  const [profile, setProfile] = useState<UserDetails | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const loadProfile = async () => {
    setLoading(true);
    setError(null);
    try {
      const me = await getCurrentUser();
      setProfile(me as UserDetails);
    } catch (e) {
      setError(parseError(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadProfile();
  }, []);

  return (
    <section className="theme-card space-y-3 rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Профиль</h2>
        <button className="interactive-chip rounded-lg border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" onClick={() => void loadProfile()}>
          Обновить
        </button>
      </div>

      {loading && <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка...</p>}
      {error && <p className="field-error">{error}</p>}

      {!loading && (
        <div className="space-y-1 text-sm text-[rgb(var(--text-primary))]">
          <p><span className="font-medium">ID:</span> {profile?.id ?? authUser?.id ?? '—'}</p>
          <p><span className="font-medium">Email:</span> {profile?.email ?? authUser?.email ?? '—'}</p>
          <p><span className="font-medium">Имя:</span> {profile?.name ?? '—'}</p>
          <p><span className="font-medium">Роль:</span> {profile?.role ?? authUser?.role ?? '—'}</p>
          <p><span className="font-medium">Последний вход:</span> {profile?.last_login ?? '—'}</p>
        </div>
      )}
    </section>
  );
};
