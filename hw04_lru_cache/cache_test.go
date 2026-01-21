package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("выталкивание элементов из-за размера очереди", func(t *testing.T) {
		// Создаем кэш на 3 элемента
		c := NewCache(3)

		// Добавляем 3 элемента
		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)
		// очередь: [c, b, a]

		// Проверяем, что все три элемента на месте
		val, ok := c.Get("a")
		require.True(t, ok)
		require.Equal(t, 1, val)

		val, ok = c.Get("b")
		require.True(t, ok)
		require.Equal(t, 2, val)

		val, ok = c.Get("c")
		require.True(t, ok)
		require.Equal(t, 3, val)

		// Добавляем 4-й элемент
		c.Set("d", 4)
		// очередь: [d, c, b]

		// Элемент 'a' должен быть вытолкнут
		val, ok = c.Get("a")
		require.False(t, ok)
		require.Nil(t, val)

		// Остальные элементы должны остаться
		val, ok = c.Get("b")
		require.True(t, ok)
		require.Equal(t, 2, val)

		val, ok = c.Get("c")
		require.True(t, ok)
		require.Equal(t, 3, val)

		val, ok = c.Get("d")
		require.True(t, ok)
		require.Equal(t, 4, val)
	})

	t.Run("выталкивание давно используемых элементов", func(t *testing.T) {
		// Создаем кэш на 3 элемента
		c := NewCache(3)

		// Добавляем 3 элемента
		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)
		// очередь: [c, b, a]

		// Обращаемся к 'a'
		val, ok := c.Get("a")
		require.True(t, ok)
		require.Equal(t, 1, val)
		// очередь: [a, c, b]

		// Обращаемся к 'b'
		val, ok = c.Get("b")
		require.True(t, ok)
		require.Equal(t, 2, val)
		// очередь: [b, a, c]

		// Обновляем значение 'c'
		wasInCache := c.Set("c", 33)
		require.True(t, wasInCache)
		// очередь: [c, b, a]

		// Добавляем 4-й элемент
		c.Set("d", 4)
		// очередь: [d, c, b]

		// Проверяем, что 'a' вытолкнут
		val, ok = c.Get("a")
		require.False(t, ok)
		require.Nil(t, val)

		// Проверяем, что остальные элементы на месте
		val, ok = c.Get("b")
		require.True(t, ok)
		require.Equal(t, 2, val)

		val, ok = c.Get("c")
		require.True(t, ok)
		require.Equal(t, 33, val)

		val, ok = c.Get("d")
		require.True(t, ok)
		require.Equal(t, 4, val)
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
