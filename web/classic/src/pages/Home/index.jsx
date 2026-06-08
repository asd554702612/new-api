/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import { API, getLogo, getSystemName, showError } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { UserContext } from '../../context/User';
import { useActualTheme, useSetTheme, useTheme } from '../../context/Theme';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { normalizeLanguage } from '../../i18n/language';
import NoticeModal from '../../components/layout/NoticeModal';
import HomeHeader from './components/HomeHeader';
import HomeHero from './components/HomeHero';
import HomeFeatures from './components/HomeFeatures';
import HomeProducts from './components/HomeProducts';
import HomePricing from './components/HomePricing';
import HomeShowcase from './components/HomeShowcase';
import HomeCTA from './components/HomeCTA';
import HomeFooter from './components/HomeFooter';
import { getHomeLandingData, PROJECT_REPOSITORY_URL } from './homeLandingData';
import './homeLanding.css';

const HOME_HEADER_OFFSET = 100;

const isUrlContent = (content) =>
  content.trim().startsWith('http://') || content.trim().startsWith('https://');

const readLocalUser = () => {
  try {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : undefined;
  } catch {
    return undefined;
  }
};

const Home = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const [userState] = useContext(UserContext);
  const actualTheme = useActualTheme();
  const theme = useTheme();
  const setTheme = useSetTheme();
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [noticeVisible, setNoticeVisible] = useState(false);
  const isMobile = useIsMobile();

  const docsLink = statusState?.status?.docs_link || '';
  const supportContactInfo = statusState?.status?.support_contact_info || '';
  const systemName = getSystemName();
  const siteLogo = getLogo();
  const localUser = readLocalUser();
  const currentUser = userState.user || localUser;
  const isAuthenticated = Boolean(currentUser);
  const isAdmin = Number(currentUser?.role || 0) >= 10;
  const dashboardPath = isAdmin ? '/console' : '/console';
  const consolePath = isAuthenticated
    ? dashboardPath
    : `/login?redirect=${encodeURIComponent(dashboardPath)}`;
  const isDark = actualTheme === 'dark';
  const isDemoSiteMode = statusState?.status?.demo_site_enabled || false;
  const demoVersion =
    isDemoSiteMode && statusState?.status?.version
      ? statusState.status.version
      : '';

  const landingData = useMemo(
    () => getHomeLandingData({ t, docsLink, systemName }),
    [docsLink, systemName, t],
  );

  const displayHomePageContent = useCallback(async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');

    try {
      const res = await API.get('/api/home_page_content');
      const { success, message, data } = res.data;

      if (success) {
        const rawContent = String(data || '');
        const content = isUrlContent(rawContent)
          ? rawContent
          : marked.parse(rawContent);

        setHomePageContent(content);
        localStorage.setItem('home_page_content', content);
      } else {
        showError(message);
        setHomePageContent(t('加载首页内容失败...'));
      }
    } catch (error) {
      showError(error.message || t('加载首页内容失败...'));
      setHomePageContent(t('加载首页内容失败...'));
    } finally {
      setHomePageContentLoaded(true);
    }
  }, [t]);

  const handleIframeLoad = useCallback(
    (event) => {
      event.currentTarget.contentWindow?.postMessage(
        { themeMode: actualTheme },
        '*',
      );
      event.currentTarget.contentWindow?.postMessage(
        { lang: i18n.language },
        '*',
      );
    },
    [actualTheme, i18n.language],
  );

  const handleLanguageChange = useCallback(
    (language) => {
      const normalizedLanguage = normalizeLanguage(language);
      i18n.changeLanguage(normalizedLanguage);
      localStorage.setItem('i18nextLng', normalizedLanguage);
    },
    [i18n],
  );

  const handleToggleTheme = useCallback(() => {
    setTheme(isDark ? 'light' : 'dark');
  }, [isDark, setTheme]);

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();

      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };

    checkNoticeAndShow();
  }, []);

  useEffect(() => {
    displayHomePageContent();
  }, [displayHomePageContent]);

  const renderDefaultLanding = () => (
    <div className='classic-home-landing-shell relative min-h-screen overflow-hidden bg-[#f6f9ff] text-slate-950 transition-colors duration-300 dark:bg-[#040816] dark:text-white'>
      <div className='pointer-events-none absolute inset-0 overflow-hidden'>
        <div className='classic-home-landing-grid absolute inset-0 opacity-60 dark:opacity-20' />
        <div className='absolute -left-20 top-20 h-[28rem] w-[28rem] rounded-full bg-blue-400/20 blur-[70px]' />
        <div className='absolute -right-10 top-40 h-[24rem] w-[24rem] rounded-full bg-cyan-300/20 blur-[70px]' />
        <div className='classic-home-landing-ribbon absolute inset-x-0 top-0 h-[36rem]' />
      </div>

      <HomeHeader
        brand={systemName}
        consoleLabel={landingData.hero.primaryAuthed}
        consolePath={consolePath}
        currentLanguage={i18n.language}
        demoVersion={demoVersion}
        isDark={isDark}
        navItems={landingData.navItems}
        onLanguageChange={handleLanguageChange}
        onToggleTheme={handleToggleTheme}
        projectUrl={PROJECT_REPOSITORY_URL}
        siteLogo={siteLogo}
        t={t}
        theme={theme}
        themeTitle={isDark ? t('切换到浅色模式') : t('切换到深色模式')}
      />

      <main className='relative z-10 px-4 pb-16 pt-[100px] sm:px-6'>
        <HomeHero
          consolePath={consolePath}
          data={landingData.hero}
          docUrl={landingData.docUrl}
          headerOffset={HOME_HEADER_OFFSET}
          primaryLabel={
            isAuthenticated
              ? landingData.hero.primaryAuthed
              : landingData.hero.primaryGuest
          }
        />
        <HomeFeatures data={landingData.features} />
        <HomeProducts data={landingData.products} />
        <HomePricing data={landingData.pricing} />
        <HomeShowcase data={landingData.showcase} />
        <HomeCTA
          consolePath={consolePath}
          data={landingData.cta}
          docUrl={landingData.docUrl}
          primaryLabel={
            isAuthenticated
              ? landingData.hero.primaryAuthed
              : landingData.hero.primaryGuest
          }
          secondaryLabel={landingData.hero.secondary}
        />
      </main>

      <HomeFooter
        brand={systemName}
        data={landingData.footer}
        docsLabel={landingData.hero.secondary}
        docUrl={landingData.docUrl}
        siteLogo={siteLogo}
        supportContactInfo={supportContactInfo}
        supportContactLabel={t('客服联系方式')}
      />
    </div>
  );

  const renderCustomContent = () => {
    if (isUrlContent(homePageContent)) {
      return (
        <iframe
          src={homePageContent.trim()}
          className='h-screen w-full border-none'
          onLoad={handleIframeLoad}
          allowFullScreen
          title={t('首页内容')}
        />
      );
    }

    return (
      <div
        className='mt-[60px]'
        dangerouslySetInnerHTML={{ __html: homePageContent }}
      />
    );
  };

  return (
    <div className='classic-page-fill classic-home-page w-full overflow-x-hidden'>
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />
      {!homePageContentLoaded ? null : homePageContent === '' ? (
        renderDefaultLanding()
      ) : (
        <div className='classic-page-fill w-full overflow-x-hidden'>
          {renderCustomContent()}
        </div>
      )}
    </div>
  );
};

export default Home;
