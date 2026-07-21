// 1. URL твоего эндпоинта
const URL = "http://localhost:8080/api/pins/create";

// 2. Твой массив с данными для наполнения
const items = [
  {
    title: "Sunset over Mountains",
    description: "A breathtaking view of a sunset over the mountains.",
    image_url:
      "https://i.pinimg.com/200x/8d/cd/03/8dcd03ff29870e5da56ee7ca5892add7.jpg",
    dominant_color: "#FFA07A",
    tags: ["nature", "sunset", "mountains"],
    width: 200,
    height: 200,
    author_id: 1,
  },
  {
    title: "Abstract Art",
    description: "An abstract painting with vibrant colors.",
    image_url:
      "https://i.pinimg.com/200x/4a/2a/a3/4a2aa3be57423f9a2285949923fcfd26.jpg",
    dominant_color: "#FFD700",
    tags: ["art", "abstract", "colorful"],
    width: 200,
    height: 200,
    author_id: 1,
  },
  {
    title: "Urban Cityscape",
    description: "A bustling city skyline at night.",
    image_url:
      "https://i.pinimg.com/200x/04/22/2c/04222cbf446453c0da3126f0acbd29e4.jpg",
    dominant_color: "#8A2BE2",
    tags: ["city", "skyline", "urban"],
    width: 200,
    height: 200,
    author_id: 1,
  },
  {
    title: "Digital Art",
    description: "An artistic depiction of a digital world.",
    image_url:
      "https://i.pinimg.com/474x/32/93/17/3293179790b97a732dead8c73e18cd1a.jpg",
    dominant_color: "#FF6347",
    tags: ["digital", "art", "technology"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Floral Design",
    description: "A delicate floral design with pastel colors.",
    image_url:
      "https://i.pinimg.com/474x/ad/14/da/ad14da8046f36f8a89044bf0c507c21f.jpg",
    dominant_color: "#FFFACD",
    tags: ["floral", "design", "pastel"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Aesthetic Landscape",
    description: "An aesthetically pleasing landscape with rolling hills.",
    image_url:
      "https://i.pinimg.com/474x/3d/64/01/3d64012ab4fdca75adf4a96847b52afd.jpg",
    dominant_color: "#FFDAB9",
    tags: ["landscape", "hills", "aesthetic"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Minimalist Design",
    description: "A minimalist design with a focus on simplicity.",
    image_url:
      "https://i.pinimg.com/474x/6a/88/16/6a8816440ccaf23320a8369105bb8be1.jpg",
    dominant_color: "#FFFFFF",
    tags: ["minimalist", "design", "simple"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Vintage Photography",
    description: "A vintage photograph of a city street.",
    image_url:
      "https://i.pinimg.com/474x/18/a6/01/18a601ce95bd5afd7edc9725db466ea2.jpg",
    dominant_color: "#C0C0C0",
    tags: ["vintage", "photography", "street"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Modern Architecture",
    description: "A modern architectural design with clean lines.",
    image_url:
      "https://i.pinimg.com/474x/dd/f0/d3/ddf0d3e2ced70f1ddd9344e36f73c3d1.jpg",
    dominant_color: "#87CEEB",
    tags: ["architecture", "modern", "clean"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Beach Scene",
    description: "A serene beach scene with palm trees.",
    image_url:
      "https://i.pinimg.com/474x/87/68/de/8768de55ec77567896dfe1b347a81043.jpg",
    dominant_color: "#FFDEAD",
    tags: ["beach", "palm trees", "serene"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Winter Landscape",
    description: "A winter landscape with snow-covered trees.",
    image_url:
      "https://i.pinimg.com/474x/84/00/e2/8400e246454feaaaf75a31007dd78a90.jpg",
    dominant_color: "#ADD8E6",
    tags: ["winter", "landscape", "snow"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Macro Photography",
    description: "A close-up macro photograph of a flower.",
    image_url:
      "https://i.pinimg.com/474x/0e/e7/d5/0ee7d5d92d8b14969467f8729c86c452.jpg",
    dominant_color: "#FFB6C1",
    tags: ["macro", "photography", "flower"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Urban Street Art",
    description: "An urban street with vibrant street art.",
    image_url:
      "https://i.pinimg.com/474x/43/30/10/433010c1d6f1ef613539f493fe88b351.jpg",
    dominant_color: "#FF4500",
    tags: ["street art", "urban", "vibrant"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Aesthetic Photography",
    description: "An aesthetically pleasing photograph of a building.",
    image_url:
      "https://i.pinimg.com/474x/69/84/ee/6984ee0b43a4268aa93cd36749dfc558.jpg",
    dominant_color: "#FFA500",
    tags: ["photography", "aesthetic", "building"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Abstract Painting",
    description: "An abstract painting with bold colors.",
    image_url:
      "https://i.pinimg.com/474x/e9/12/cb/e912cb30856f32c7913bda96ee76c763.jpg",
    dominant_color: "#FFD700",
    tags: ["painting", "abstract", "bold"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Modern Furniture",
    description: "A modern piece of furniture with sleek design.",
    image_url:
      "https://i.pinimg.com/474x/87/f3/1c/87f31c6862c8951cdc774acfb2193cc4.jpg",
    dominant_color: "#B0C4DE",
    tags: ["furniture", "modern", "sleek"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Nature Photography",
    description: "A stunning nature photograph of a waterfall.",
    image_url:
      "https://i.pinimg.com/474x/f0/ce/04/f0ce047a74150921ec600d98603cb068.jpg",
    dominant_color: "#2F4F4F",
    tags: ["photography", "nature", "waterfall"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Geometric Art",
    description: "A geometric art piece with clean lines.",
    image_url:
      "https://i.pinimg.com/474x/e9/b2/01/e9b201b6a9989bc8a74f8320a99a54cf.jpg",
    dominant_color: "#8B4513",
    tags: ["art", "geometric", "clean"],
    width: 474,
    height: 474,
    author_id: 1,
  },
  {
    title: "Urban Life",
    description: "A busy urban street scene.",
    image_url:
      "https://i.pinimg.com/474x/b3/83/b3/b383b3757b37ed0babfe05682d31fe61.jpg",
    dominant_color: "#FFA07A",
    tags: ["urban", "life", "busy"],
    width: 474,
    height: 474,
    author_id: 1,
  },
];

// Функция для задержки (аналог time.Sleep в Go)
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

async function runSeeder() {
  console.log(`Начинаем отправку ${items.length} элементов...`);

  for (let i = 0; i < items.length; i++) {
    const item = items[i];

    try {
      // Отправляем POST запрос
      const response = await fetch(URL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Cookie: "session=LEVXuih4zIFEOpPZdwE7FILDuOvlP20vYwIEsoduUzM",
          // Если нужна авторизация, раскомментируй строку ниже:
          // 'Authorization': 'Bearer ТВОЙ_ТОКЕН'
        },
        body: JSON.stringify(item),
      });

      if (response.ok) {
        console.log(`[${i}] Успешно отправлено: "${item.title}"`);
      } else {
        console.error(
          `[${i}] Бэкенд вернул ошибку: ${response.status} ${response.statusText}`,
        );
      }
    } catch (error) {
      console.error(`[${i}] Ошибка запроса:`, error.message);
    }

    // Небольшая пауза в 100 мс между запросами
    await sleep(100);
  }

  console.log("Готово!");
}

runSeeder();
