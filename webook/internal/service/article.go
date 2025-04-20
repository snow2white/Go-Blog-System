package service

import (
	"basic-go/webook/internal/domain"
	"basic-go/webook/internal/events/article"
	"basic-go/webook/internal/repository"
	"basic-go/webook/pkg/logger"
	"context"
	"fmt"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer

	// V1 写法专用
	// readerRepo repository.ArticleReaderRepository
	// authorRepo repository.ArticleAuthorRepository
	l logger.LoggerV1
}

// func NewArticleServiceV1(
// 	readerRepo repository.ArticleReaderRepository,
// 	authorRepo repository.ArticleAuthorRepository, l logger.LoggerV1) *articleService {
// 	return &articleService{
// 		readerRepo: readerRepo,
// 		authorRepo: authorRepo,
// 		l:          l,
// 	}
// }

func NewArticleService(repo repository.ArticleRepository,
	producer article.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

func (a *articleService) GetPubById(ctx context.Context, id, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)
	go func() {
		if err == nil {
			// 在这里发一个消息
			fmt.Println("Aid: ", id, "   uid:", uid)
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(err))
			}
		}
	}()
	return res, err
}

// func (a *articleService) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
// 	return a.repo.GetPubById(ctx, id)
// }

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (a *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	// art.Status = domain.ArticleStatusPublished
	// return a.repo.Sync(ctx, art)
	art.Status = domain.ArticleStatusPublished
	res, err := a.repo.Sync(ctx, art)
	fmt.Println("res: ", res)
	go func() {
		if err == nil {
			// 在这里发一个消息
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Aid: res,
				Uid: art.Author.Id,
			})
			fmt.Println("publish,art.Id: ", art.Id)
			if er != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", art.Id),
					logger.Int64("uid", art.Author.Id),
					logger.Error(err))
			}
		}
	}()
	return res, err
}

// 具备重试机制的发布
// func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
// 	// 想到这里要先操作制作库
// 	// 这里操作线上库
// 	var (
// 		id  = art.Id
// 		err error
// 	)
// 	if art.Id > 0 {
// 		err = a.authorRepo.Update(ctx, art)
// 	} else {
// 		id, err = a.authorRepo.Create(ctx, art)
// 	}
// 	if err != nil {
// 		return 0, err
// 	}
// 	art.Id = id
// 	for i := 0; i < 3; i++ {
// 		// 我可能线上库已经有数据了
// 		// 也可能没有
// 		err = a.readerRepo.Save(ctx, art)
// 		if err != nil {
// 			// 多接入一些 tracing 的工具
// 			a.l.Error("保存到制作库成功但是到线上库失败",
// 				logger.Int64("aid", art.Id),
// 				logger.Error(err))
// 		} else {
// 			return id, nil
// 		}
// 	}
// 	a.l.Error("保存到制作库成功但是到线上库失败，重试耗尽",
// 		logger.Int64("aid", art.Id),
// 		logger.Error(err))
// 	return id, errors.New("保存到线上库失败，重试次数耗尽")
// }

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}
