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

import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Check, ChevronDown, Menu, Moon, Sun, X } from 'lucide-react';

const localeOptions = [
  { code: 'zh-CN', name: '简体中文', shortCode: 'ZH', flag: '🇨🇳' },
  { code: 'zh-TW', name: '繁體中文', shortCode: 'ZH', flag: '🇭🇰' },
  { code: 'en', name: 'English', shortCode: 'EN', flag: '🇺🇸' },
  { code: 'fr', name: 'Français', shortCode: 'FR', flag: '🇫🇷' },
  { code: 'ja', name: '日本語', shortCode: 'JA', flag: '🇯🇵' },
  { code: 'ru', name: 'Русский', shortCode: 'RU', flag: '🇷🇺' },
  { code: 'vi', name: 'Tiếng Việt', shortCode: 'VI', flag: '🇻🇳' },
];

const getCurrentLocale = (currentLang) => {
  const normalized = currentLang === 'zh' ? 'zh-CN' : currentLang;
  return (
    localeOptions.find((locale) => locale.code === normalized) ||
    localeOptions[0]
  );
};

const HomeLocaleSwitcher = ({ currentLang, onLanguageChange }) => {
  const [open, setOpen] = useState(false);
  const currentLocale = getCurrentLocale(currentLang);

  useEffect(() => {
    if (!open) {
      return undefined;
    }

    const handleClick = (event) => {
      if (!event.target.closest('[data-home-locale-switcher]')) {
        setOpen(false);
      }
    };

    document.addEventListener('click', handleClick);
    return () => document.removeEventListener('click', handleClick);
  }, [open]);

  return (
    <div className='relative' data-home-locale-switcher>
      <button
        type='button'
        onClick={() => setOpen((value) => !value)}
        className='flex items-center gap-1.5 rounded-lg px-2 py-1.5 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'
        title={currentLocale.name}
      >
        <span className='text-base'>{currentLocale.flag}</span>
        <span className='hidden sm:inline'>
          {currentLocale.shortCode.toUpperCase()}
        </span>
        <ChevronDown
          size={12}
          className={`text-gray-400 transition-transform duration-200 ${
            open ? 'rotate-180' : ''
          }`}
        />
      </button>

      {open ? (
        <div className='absolute right-0 z-50 mt-1 w-32 overflow-hidden rounded-lg border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800'>
          {localeOptions.map((locale) => {
            const selected = locale.code === currentLocale.code;

            return (
              <button
                key={locale.code}
                type='button'
                onClick={() => {
                  onLanguageChange(locale.code);
                  setOpen(false);
                }}
                className={`flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-dark-700 ${
                  selected
                    ? 'bg-primary-50 text-primary-600 dark:bg-primary-900/20 dark:text-primary-400'
                    : ''
                }`}
              >
                <span className='text-base'>{locale.flag}</span>
                <span>{locale.name}</span>
                {selected ? (
                  <Check size={16} className='ml-auto text-primary-500' />
                ) : null}
              </button>
            );
          })}
        </div>
      ) : null}
    </div>
  );
};

const HomeHeader = (props) => {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 10);
    };

    handleScroll();
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const renderNavLink = (item, keyPrefix = '') => (
    <a
      key={`${keyPrefix}${item.label}`}
      href={item.href}
      target={item.external ? '_blank' : undefined}
      rel={item.external ? 'noopener noreferrer' : undefined}
      className='text-[13px] font-medium text-zinc-500 transition-colors hover:text-black dark:text-slate-300 dark:hover:text-white'
      onClick={() => setMobileMenuOpen(false)}
    >
      {item.label}
    </a>
  );

  return (
    <header
      className={`fixed inset-x-0 top-0 z-50 transition-all duration-300 ${
        scrolled
          ? 'border-b border-zinc-100/80 bg-white/80 py-3 backdrop-blur-xl dark:border-slate-800/80 dark:bg-slate-950/85'
          : 'bg-transparent py-5'
      }`}
    >
      <nav className='mx-auto max-w-[1400px] px-6'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-10'>
            <a href='#top' className='flex items-center gap-3'>
              <div className='flex h-8 w-8 items-center justify-center overflow-hidden rounded-full bg-white ring-1 ring-slate-200/70 dark:bg-slate-900/90 dark:ring-slate-700/80'>
                <img
                  src={props.siteLogo || '/favicon.ico'}
                  alt={props.brand}
                  className='h-full w-full object-contain p-1'
                />
              </div>
              <span className='text-lg font-bold tracking-tighter text-black dark:text-white'>
                {props.brand}
              </span>
            </a>

            <div className='hidden items-center space-x-10 lg:flex'>
              {props.navItems.map((item) => renderNavLink(item))}
            </div>
          </div>

          <div className='hidden items-center space-x-6 lg:flex'>
            <HomeLocaleSwitcher
              currentLang={props.currentLanguage}
              onLanguageChange={props.onLanguageChange}
            />

            <button
              type='button'
              onClick={props.onToggleTheme}
              className='inline-flex h-9 w-9 items-center justify-center rounded-full border border-zinc-200 bg-white/70 text-zinc-500 transition-colors hover:text-black dark:border-slate-700/80 dark:bg-slate-900/90 dark:text-slate-300 dark:hover:text-white'
              title={props.themeTitle}
              aria-label={props.themeTitle}
            >
              {props.isDark ? <Sun size={16} /> : <Moon size={16} />}
            </button>

            {props.demoVersion ? (
              <a
                href={props.projectUrl}
                target='_blank'
                rel='noopener noreferrer'
                className='rounded-full border border-zinc-200 bg-white/70 px-4 py-2 text-[13px] font-medium text-zinc-600 transition-colors hover:text-black dark:border-slate-700/80 dark:bg-slate-900/90 dark:text-slate-300 dark:hover:text-white'
              >
                {props.demoVersion}
              </a>
            ) : null}

            <Link
              to={props.consolePath}
              className='rounded-full bg-black px-6 py-2 text-[13px] font-medium text-white shadow-sm transition-all hover:bg-zinc-800 active:scale-95 dark:border dark:border-cyan-400/30 dark:bg-cyan-300 dark:text-slate-950 dark:hover:bg-cyan-200'
            >
              {props.consoleLabel}
            </Link>
          </div>

          <div className='lg:hidden'>
            <button
              type='button'
              onClick={() => setMobileMenuOpen((open) => !open)}
              className='rounded-full border border-zinc-200 bg-white/80 p-2 text-zinc-600 transition-colors hover:text-black dark:border-slate-700/80 dark:bg-slate-900/90 dark:text-slate-300 dark:hover:text-white'
              aria-label={
                mobileMenuOpen ? props.t('关闭菜单') : props.t('打开菜单')
              }
            >
              {mobileMenuOpen ? <X size={20} /> : <Menu size={20} />}
            </button>
          </div>
        </div>
      </nav>

      {mobileMenuOpen ? (
        <div className='absolute left-0 right-0 top-full border-b border-zinc-100 bg-white shadow-xl dark:border-slate-800/80 dark:bg-slate-950 lg:hidden'>
          <div className='flex flex-col space-y-4 px-6 py-4'>
            {props.navItems.map((item) => renderNavLink(item, 'mobile-'))}
            <div className='flex items-center gap-3 pt-4'>
              <HomeLocaleSwitcher
                currentLang={props.currentLanguage}
                onLanguageChange={props.onLanguageChange}
              />
              <button
                type='button'
                onClick={props.onToggleTheme}
                className='inline-flex h-10 w-10 items-center justify-center rounded-full border border-zinc-200 bg-white text-zinc-500 dark:border-slate-700/80 dark:bg-slate-900/90 dark:text-slate-300'
                title={props.themeTitle}
                aria-label={props.themeTitle}
              >
                {props.isDark ? <Sun size={16} /> : <Moon size={16} />}
              </button>
            </div>
            <Link
              to={props.consolePath}
              className='w-full rounded-full bg-black px-6 py-3 text-center text-[15px] font-medium text-white dark:border dark:border-cyan-400/30 dark:bg-cyan-300 dark:text-slate-950'
              onClick={() => setMobileMenuOpen(false)}
            >
              {props.consoleLabel}
            </Link>
          </div>
        </div>
      ) : null}
    </header>
  );
};

export default HomeHeader;
